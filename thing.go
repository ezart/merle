// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

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
}

type Thing struct {
	thinger     Thinger
	cfg         *ThingConfig
	assets      *ThingAssets
	status      string
	id          string
	model       string
	name        string
	startupTime time.Time
	bus         *bus
	tunnel      *tunnel
	isWeber     bool
	web         *web
	isBridge    bool
	bridge      *bridge
	isPrime     bool
	primePort   *port
	primeId     string
	log         *log.Logger
}

// NewThing will return a Thing built from a Thinger and a ThingConfig.  E.g.
//
//	func main() {
//		var cfg merle.ThingConfig
//	
//		fooer := foo.NewFooer()
//		thing := merle.NewThing(fooer, &cfg)
//	
//		log.Fatalln(thing.Run())
//	}
//
func NewThing(thinger Thinger, cfg *ThingConfig) *Thing {

	id := cfg.Thing.Id
	isPrime := cfg.Thing.Prime

	if !isPrime {
		if id == "" {
			id = defaultId()
			log.Println("Defaulting ID to", id)
		}
	}

	prefix := "[" + id + "] "

	t := &Thing{
		thinger:     thinger,
		cfg:         cfg,
		status:      "online",
		id:          id,
		model:       cfg.Thing.Model,
		name:        cfg.Thing.Name,
		startupTime: time.Now(),
		isPrime:     isPrime,
		log:         log.New(os.Stderr, prefix, 0),
	}

	t.bus = newBus(t, 10, thinger.Subscribers())

	t.tunnel = newTunnel(t.id, cfg.Mother.Host, cfg.Mother.User,
		cfg.Mother.Key, cfg.Thing.PortPrivate, cfg.Mother.PortPrivate)

	_, t.isWeber = t.thinger.(Weber)
	if t.isWeber {
		t.assets = t.thinger.(Weber).Assets()
		t.web = newWeb(t, cfg.Thing.PortPublic, cfg.Thing.PortPublicTLS,
			cfg.Thing.PortPrivate, cfg.Thing.User, t.assets)
		t.setAssetsDir(t)
	}

	_, t.isBridge = t.thinger.(Bridger)
	if t.isBridge {
		t.bridge = newBridge(t)
	}

	if t.isPrime {
		t.web.handlePrimePortId()
		t.primePort = newPort(t, cfg.Thing.PortPrime, t.primeAttach)
	}

	t.bus.subscribe(GetIdentity, t.getIdentity)

	return t
}

func (t *Thing) getIdentity(p *Packet) {
	resp := MsgIdentity{
		Msg:         ReplyIdentity,
		Status:      t.status,
		Id:          t.id,
		Model:       t.model,
		Name:        t.name,
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
	if t.isWeber {
		t.web.start()
	}

	t.tunnel.start()

	if t.isBridge {
		t.bridge.start()
	}

	msg := struct{ Msg string }{Msg: CmdRun}
	t.bus.receive(newPacket(t.bus, nil, &msg))

	if t.isBridge {
		t.bridge.stop()
	}

	t.tunnel.stop()

	if t.isWeber {
		t.web.stop()
	}

	t.bus.close()

	return fmt.Errorf("CmdRun didn't run forever")
}

func (t *Thing) Run() error {

	re := regexp.MustCompile("^[a-zA-Z0-9_]*$")

	if !re.MatchString(t.id) {
		return fmt.Errorf("Id must contain only alphanumeric or underscore characters")
	}
	if !re.MatchString(t.model) {
		return fmt.Errorf("Model must contain only alphanumeric or underscore characters")
	}
	if !re.MatchString(t.name) {
		return fmt.Errorf("Name must contain only alphanumeric or underscore characters")
	}

	switch {
	case t.isPrime:
		return t.runPrime()
	default:
		return t.run()
	}
}

func (t *Thing) runOnPort(p *port) error {
	var name = fmt.Sprintf("port:%d", p.port)
	var sock = newWebSocket(name, p.ws)
	var pkt = newPacket(t.bus, sock, nil)
	var err error

	t.log.Printf("Websocket opened [%s]", name)

	t.bus.plugin(sock)

	msg := struct{ Msg string }{Msg: CmdRunPrime}
	t.bus.receive(pkt.Marshal(&msg))

	for {
		// new pkt for each rcv
		var pkt = newPacket(t.bus, sock, nil)

		pkt.msg, err = p.readMessage()
		if err != nil {
			t.log.Printf("Websocket closed [%s]", name)
			break
		}
		t.bus.receive(pkt)
	}

	t.bus.unplug(sock)

	return err
}

func (t *Thing) setAssetsDir(child *Thing) {
	t.web.staticFiles(child.assets.Dir, "/" + child.id + "/assets/")
}
