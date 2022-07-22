// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"log"
	"os"
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

// All Things implement the Thinger interface.
//
// To be a Thinger, the Thing must implement the two methods Subscribers() and Assets():
//
//	type thing struct {}
//	func (t *thing) Subscribers() merle.Subscribers { ... }
//	func (t *thing) Assets() *merle.ThingAssets { ... }
//
type Thinger interface {

	// Map of Thing's subscribers, keyed by message.  On Packet receipt, a
	// subscriber is looked up by Packet message.  If there is a match, the
	// subscriber callback is called.  If no subscribers match the received
	// message, the "default" subscriber matches.  If still no matches, the
	// Packet is not handled.  If the callback is nil, the Packet is
	// (silently) dropped.  Here is an example of a subscriber map:
	//
	//	func (t *thing) Subscribers() merle.Subscribers {
	//		return merle.Subscribers{
	//			merle.CmdRun:     t.run,
	//			merle.GetState:   t.getState,
	//			merle.ReplyState: t.saveState,
	//			"SpamUpdate":     t.update,
	//			"SpamTimer":      nil,         // silent drop
	//		}
	//	}
	//
	Subscribers() Subscribers

	// Thing's web server assets.
	Assets() *ThingAssets
}

// Thing made from a Thinger.
type Thing struct {
	// Thing's configuration
	Cfg         ThingConfig
	thinger     Thinger
	assets      *ThingAssets
	id          string
	model       string
	name        string
	online      bool
	startupTime time.Time
	bus         *bus
	tunnel      *tunnel
	web         *web
	isBridge    bool
	bridge      *bridge
	isPrime     bool
	primePort   *port
	primeSock   *webSocket
	primeId     string
	bridgeSock  *wireSocket
	childSock   *wireSocket
	log         *log.Logger
}

// NewThing returns a Thing built from a Thinger.
//
//	type thing struct {
//		// Implements Thinger interface
//	}
//
//	func main() {
//		merle.NewThing(&thing{}).Run()
//	}
//
func NewThing(thinger Thinger) *Thing {
	return &Thing{
		Cfg:     defaultCfg,
		thinger: thinger,
		assets:  thinger.Assets(),
	}
}

func (t *Thing) getIdentity(p *Packet) {
	resp := MsgIdentity{
		Msg:         ReplyIdentity,
		Id:          t.id,
		Model:       t.model,
		Name:        t.name,
		Online:      t.online,
		StartupTime: t.startupTime,
	}
	p.Marshal(&resp).Reply()
}

func (t *Thing) run() error {

	t.online = true

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

	if !validId(t.Cfg.Id) {
		return fmt.Errorf("Id must contain only alphanumeric or underscore characters")
	}
	if !validModel(t.Cfg.Model) {
		return fmt.Errorf("Model must contain only alphanumeric or underscore characters")
	}
	if !validName(t.Cfg.Name) {
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
		t.tunnel = newTunnel(t, t.Cfg.MotherHost,
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

// Run Thing.  An error is returned if Run() fails.  Configure Thing before
// running.
//
//	func main() {
//		thing := merle.NewThing(&thing{})
//		thing.Cfg.PortPublic = 80  // run public web server on port :80
//		log.Fatalln(thing.Run())
//	}
//
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
