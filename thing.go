// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"regexp"
	"time"
)

type ThingAssets struct {

	// Directory on file system for Thing's assets (html, css, js, etc)
	// This is an absolute or relative directory.  If relative, it's
	// relative to the Thing's binary path.
	AssetsDir string

	// Path to Thing's HTML template file, relative to AssetsDir.
	HtmlTemplate string

	// HtmlTemplateText is text passed in lieu of a template file.
	// HtmlTemplateText takes priority over HtmlTemplate, if both are
	// present.
	HtmlTemplateText string
}

// All Things implement this interface.
//
// A Thing's subscribers handle incoming messages.  The collection of message
// handlers comprise the Thing's "model".
//
// Minimally, the Thing should subscibe to the CmdRun message as CmdRun is the
// Thing's main loop.  This loop should run forever.  It is an error for CmdRun
// to end.  The main loop initializes the Thing's resources and asynchronously
// monitors and updates those resources.
//
// Here's an example of a CmdRun handler which initializes some hardware
// resources and then (asyncrounously) polls for hardware updates.
//
//	func (t *thing) run(p *merle.Packet) {
//
//		// Initialize hardware
//
//		t.adaptor = raspi.NewAdaptor()
//		t.adaptor.Connect()
//
//		t.led = gpio.NewLedDriver(t.adaptor, "11")
//		t.led.Start()
//
//		// Every second update hardware and send
//		// notifications
//
//		ticker := time.NewTicker(time.Second)
//
//		t.sendLedState(p)
//
//		for {
//			select {
//			case <-ticker.C:
//				t.toggle()
//				t.sendLedState(p)
//			}
//		}
//	}
//
// The Packet passed in can be used repeatably to send notifications.  Here,
// the Packet message is updated to broadcast the hardware state to listeners.
//
//	func (t *thing) sendLedState(p *merle.Packet) {
//		spam := spamLedState{
//			Msg:   "SpamLedState",
//			State: t.state(),
//		}
//		p.Marshal(&spam).Broadcast()
//	}
//
type Thinger interface {

	// Map of Thing's subscribers, keyed by message.  On packet receipt, a
	// subscriber is looked up by packet message.  If there is a match, the
	// subscriber callback is called.  If no subscribers match the received
	// message, the "default" subscriber matches.  If still no matches, the
	// packet is not handled.  If the callback is nil, the packet is
	// (silently) dropped.  Here is an example of a subscriber map:
	//
	//	func (t *thing) Subscribers() merle.Subscribers {
	//		return merle.Subscribers{
	//			merle.CmdRun: t.run,
	//			"GetState": t.getState,
	//			"ReplyState": t.saveState,
	//			"SpamUpdate": t.update,
	//			"SpamTimer": nil,         // silent drop
	//		}
	//	}
	Subscribers() Subscribers

	// blah
	Assets() *ThingAssets
}

type Thing struct {
	Cfg         ThingConfig
	thinger     Thinger
	assets      *ThingAssets
	id          string
	model       string
	name        string
	startupTime time.Time
	connected   bool
	bus         *bus
	tunnel      *tunnel
	web         *web
	templ       *template.Template
	templErr    error
	isBridge    bool
	bridge      *bridge
	isPrime     bool
	primePort   *port
	primeId     string
	bridgeSock  *wireSocket
	childSock   *wireSocket
	log         *log.Logger
}

// NewThing returns a Thing built from a Thinger.
func NewThing(thinger Thinger) *Thing {
	return &Thing{
		Cfg: defaultCfg,
		thinger: thinger,
		assets: thinger.Assets(),
	}
}

func (t *Thing) getIdentity(p *Packet) {
	resp := MsgIdentity{
		Msg:         ReplyIdentity,
		Id:          t.id,
		Model:       t.model,
		Name:        t.name,
		Connected:   t.connected,
		StartupTime: t.startupTime,
	}
	p.Marshal(&resp).Reply()
}

func (t *Thing) getChild(id string) *Thing {
	if !t.isBridge {
		return nil
	}
	return t.bridge.getChild(id)
}

func (t *Thing) run() error {

	t.connected = true

	// Force receipt of CmdInit msg
	msg := Msg{Msg: CmdInit}
	t.bus.receive(newPacket(t.bus, nil, &msg))

	// After CmdInit, It's safe now to handle html and ws requests.
	// (CmdInit initializes Thing's state, so it's safe to receive
	// GetState, even if that happens before CmdRun).

	t.web.public.start()
	t.web.private.start()

	t.tunnel.start()

	if t.isBridge {
		t.bridge.start()
	}

	// Force receipt of CmdRun msg
	msg = Msg{Msg: CmdRun}
	t.bus.receive(newPacket(t.bus, nil, &msg))

	// Thing should wait forever in CmdRun handler, but just
	// in case CmdRun handler exits, tear stuff down...

	if t.isBridge {
		t.bridge.stop()
	}

	t.tunnel.stop()

	t.web.private.stop()
	t.web.public.stop()

	return fmt.Errorf("CmdRun didn't run forever")
}

func (t *Thing) build(full bool) error {

	re := regexp.MustCompile("^[a-zA-Z0-9_]*$")

	if !re.MatchString(t.Cfg.Id) {
		return fmt.Errorf("Id must contain only alphanumeric or underscore characters")
	}
	if !re.MatchString(t.Cfg.Model) {
		return fmt.Errorf("Model must contain only alphanumeric or underscore characters")
	}
	if !re.MatchString(t.Cfg.Name) {
		return fmt.Errorf("Name must contain only alphanumeric or underscore characters")
	}

	id := t.Cfg.Id
	if !t.Cfg.IsPrime && id == "" {
		id = defaultId()
	}

	prefix := "[" + id + "] "
	t.log = log.New(os.Stderr, prefix, 0)

	t.id = id
	t.model = t.Cfg.Model
	t.name = t.Cfg.Name
	t.startupTime = time.Now()
	t.isPrime = t.Cfg.IsPrime

	t.bus = newBus(t, t.Cfg.MaxConnections, t.thinger.Subscribers())

	t.bus.subscribe(GetIdentity, t.getIdentity)

	t.setHtmlTemplate()

	if full {
		t.tunnel = newTunnel(t.log, t.id, t.Cfg.MotherHost,
			t.Cfg.MotherUser, t.Cfg.PortPrivate,
			t.Cfg.MotherPortPrivate)

		t.web = newWeb(t, t.Cfg.PortPublic, t.Cfg.PortPublicTLS,
			t.Cfg.PortPrivate, t.Cfg.User)
		t.setAssetsDir(t)

		_, t.isBridge = t.thinger.(Bridger)
		if t.isBridge {
			t.bridge = newBridge(t, t.Cfg.BridgePortBegin,
				t.Cfg.BridgePortEnd)
		}

		if t.isPrime {
			t.web.handlePrimePortId()
			t.primePort = newPort(t, t.Cfg.PortPrime, t.primeAttach)
		}
	}

	prime := ""
	if t.isPrime {
		prime = "[Thing Prime] "
	}

	t.log.Printf("%sModel: \"%s\", Name: \"%s\"", prime, t.model, t.name)

	return nil
}

func (t *Thing) Run() error {
	err := t.build(true)
	if err != nil {
		return err
	}

	switch {
	case t.isPrime:
		return t.primeRun()
	default:
		return t.run()
	}
}

func (t *Thing) setAssetsDir(child *Thing) {
	t.web.staticFiles(child)
}
