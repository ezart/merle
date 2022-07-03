// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
)

// BridgeThingers is a map of functions which can generate Thingers, keyed by a
// regular expression (re) of the form: id:model:name specifying which Things
// can attach to the bridge.
type BridgeThingers map[string]func() Thinger

// A Thing implementing the Bridger interface is a Bridge
type Bridger interface {

	// Map of Thingers supported by Bridge.  Map keyed by a regular
	// expression (re) of the form: id:model:name specifying which Things
	// can attach to the bridge. E.g.:
	//
	//	return merle.Thingers{
	//		".*:blink:.*": blink.NewBlinker,
	//		".*:gps:.*": gps.NewGps,
	//	}
	//
	// In this example, "01234:blink:blinky" would match the first entry.
	// "8888:foo:bar" would not match either entry and would not attach.
	BridgeThingers() BridgeThingers

	// List of subscribers on Bridge bus.  All packets from all connected
	// Things (children) are forwarded to the Bridge bus and tested against
	// the subscribers.
	BridgeSubscribers() Subscribers
}

// Children are the Things connected to the bridge, map keyed by Child Id
type children map[string]*Thing

// Bridge backing struct
type bridge struct {
	thing    *Thing
	thingers BridgeThingers
	children children
	bus      *bus
	ports    *ports
}

func newBridge(thing *Thing, portBegin, portEnd uint) *bridge {
	bridger := thing.thinger.(Bridger)

	b := &bridge{
		thing:    thing,
		thingers: bridger.BridgeThingers(),
		children: make(children),
		bus:      newBus(thing, thing.Cfg.MaxConnections,
			bridger.BridgeSubscribers()),
	}

	b.ports = newPorts(thing, portBegin, portEnd, b.bridgeAttach)
	b.thing.web.handleBridgePortId()

	return b
}

func (b *bridge) getChild(id string) *Thing {
	return b.children[id]
}

func (b *bridge) newChild(id, model, name string) (*Thing, error) {
	var thinger Thinger

	// TODO think about if it makes sense to allow you to be your own Mother?
	// TODO (Probably not)
	if b.thing.id == id {
		return nil, fmt.Errorf("Sorry, you can't be your own Mother")
	}

	spec := id + ":" + model + ":" + name

	for key, f := range b.thingers {
		match, err := regexp.MatchString(key, spec)
		if err != nil {
			return nil, fmt.Errorf("Thinger regexp error: %s", err)
		}
		if match {
			if f != nil {
				thinger = f()
			}
			break
		}
	}

	if thinger == nil {
		return nil, fmt.Errorf("No Thinger matched [%s], not attaching", spec)
	}

	child := NewThing(thinger)

	child.Cfg.Id = id
	child.Cfg.Model = model
	child.Cfg.Name = name
	child.Cfg.IsPrime = true

	err := child.build(false)
	if err != nil {
		return nil, err
	}

	b.thing.setAssetsDir(child)

	return child, nil
}

func (b *bridge) sendStatus(child *Thing) {
	msg := MsgEventStatus{Msg: EventStatus, Id: child.id, Online: child.online}
	b.thing.bus.receive(newPacket(b.thing.bus, nil, &msg))
	newPacket(child.bus, child.primeSock, &msg).Broadcast()
}

func (b *bridge) bridgeReady(child *Thing) {
	child.bridgeSock = newWireSocket("bridge sock", b.bus, nil)
	child.childSock = newWireSocket("child sock", child.bus, child.bridgeSock)
	child.bridgeSock.opposite = child.childSock

	b.bus.plugin(child.childSock)
	child.bus.plugin(child.bridgeSock)

	child.online = true
	b.sendStatus(child)
}

func (b *bridge) bridgeCleanup(child *Thing) {
	child.online = false
	b.sendStatus(child)

	child.bus.unplug(child.bridgeSock)
	b.bus.unplug(child.childSock)
}

func (b *bridge) bridgeAttach(p *port, msg *MsgIdentity) error {
	var err error

	child := b.getChild(msg.Id)

	if child == nil {
		child, err = b.newChild(msg.Id, msg.Model, msg.Name)
		if err != nil {
			return errors.Wrap(err, "Bridge attach creating new child")
		}
		b.children[msg.Id] = child
	} else {
		if child.model != msg.Model {
			return fmt.Errorf("Bridge attach model mismatch")
		}
		if child.name != msg.Name {
			return fmt.Errorf("Bridge attach name mismatch")
		}
	}

	child.primePort = p
	child.startupTime = msg.StartupTime

	return child.runOnPort(p, b.bridgeReady, b.bridgeCleanup)
}

func (b *bridge) start() {
	if err := b.ports.start(); err != nil {
		b.thing.log.Println("Starting bridge error:", err)
	}
}

func (b *bridge) stop() {
	b.ports.stop()
	b.bus.close()
}

// Wire socket
type wireSocket struct {
	name     string
	flags    uint32
	bus      *bus
	opposite *wireSocket
}

func newWireSocket(name string, bus *bus, opposite *wireSocket) *wireSocket {
	return &wireSocket{name: name, flags: sock_flag_bcast,
		bus: bus, opposite: opposite}
}

func (s *wireSocket) Send(p *Packet) error {
	s.bus.receive(p.clone(s.bus, s.opposite))
	return nil
}

func (s *wireSocket) Close() {
}

func (s *wireSocket) Name() string {
	return s.name
}

func (s *wireSocket) Flags() uint32 {
	return s.flags
}

func (s *wireSocket) SetFlags(flags uint32) {
	s.flags = flags
}

func (s *wireSocket) Src() string {
	return s.bus.thing.id
}
