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

func newBridge(thing *Thing) *bridge {
	bridger := thing.thinger.(Bridger)

	b := &bridge{
		thing:    thing,
		thingers: bridger.BridgeThingers(),
		children: make(children),
		bus:      newBus(thing, 10, bridger.BridgeSubscribers()),
	}

	b.ports = newPorts(thing, thing.cfg.Bridge.PortBegin,
		thing.cfg.Bridge.PortEnd, b.bridgeAttach)

	b.thing.bus.subscribe(GetChildren, b.getChildren)
	b.thing.web.handleBridgePortId()

	return b
}

func (b *bridge) getChild(id string) *Thing {
	return b.children[id]
}

func (b *bridge) changeStatus(child *Thing, sock *wireSocket, status string) {
	child.status = status

	spam := MsgSpamStatus{
		Msg:    SpamStatus,
		Id:     child.id,
		Model:  child.model,
		Name:   child.name,
		Status: child.status,
	}
	newPacket(b.thing.bus, nil, &spam).Broadcast()
	b.bus.receive(newPacket(b.bus, sock, &spam))
}

func (b *bridge) runChild(p *port, child *Thing) {
	bridgeSock := newWireSocket("bridge sock", b.bus, nil)
	childSock := newWireSocket("child sock", child.bus, bridgeSock)
	bridgeSock.opposite = childSock

	b.bus.plugin(childSock)
	child.bus.plugin(bridgeSock)

	b.changeStatus(child, childSock, "online")
	child.runOnPort(p)
	b.changeStatus(child, childSock, "offline")

	child.bus.unplug(bridgeSock)
	b.bus.unplug(childSock)
}

func (b *bridge) newChild(id, model, name string) (*Thing, error) {
	var thinger Thinger
	var cfg ThingConfig

	// TODO think about if it makes sense to allow you to be your own Mother?
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

	cfg.Thing.Id = id
	cfg.Thing.Model = model
	cfg.Thing.Name = name

	child := NewThing(thinger, &cfg)

	b.thing.setAssetsDir(child)

	return child, nil
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

	child.startupTime = msg.StartupTime

	b.runChild(p, child)

	return nil
}

func (b *bridge) getChildren(p *Packet) {
	resp := MsgChildren{Msg: ReplyChildren}
	for _, child := range b.children {
		resp.Children = append(resp.Children, MsgChild{"", child.id,
			child.model, child.name, child.status})
	}
	p.Marshal(&resp).Reply()
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
