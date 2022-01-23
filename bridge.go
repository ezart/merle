package merle

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
)

// A Thing implementing the Bridger interface is a bridge
type Bridger interface {
	BridgeThingers() Thingers
	// List of subscribers on bridge bus.  All packets from all connected
	// things (children) are forwarded to the bridge bus and tested against
	// these subscribers.
	BridgeSubscribers() Subscribers
}

// Children are the Things connected to the bridge, map keyed by Child Id
type children map[string]*Thing

// Bridge backing struct
type bridge struct {
	thing    *Thing
	thingers Thingers
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
		thing.cfg.Bridge.PortEnd, b.attachCb)

	b.thing.bus.subscribe("_GetOnlyChild", b.getOnlyChild)
	b.thing.bus.subscribe("_GetChildren", b.getChildren)

	b.thing.private.handleFunc("/port/{id}", b.getPort)

	return b
}

func (b *bridge) getChild(id string) *Thing {
	if child, ok := b.children[id]; ok {
		return child
	}
	return nil
}

type SpamStatus struct {
	Msg    string
	Id     string
	Model  string
	Name   string
	Status string
}

func (b *bridge) changeStatus(child *Thing, sock *wireSocket, status string) {
	child.status = status

	spam := SpamStatus{
		Msg:    "_SpamStatus",
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
	child.runInBridge(p)
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

	return child, nil
}

func (b *bridge) attachCb(p *port, msg *msgIdentity) error {
	var err error

	child := b.getChild(msg.Id)

	if child == nil {
		child, err = b.newChild(msg.Id, msg.Model, msg.Name)
		if err != nil {
			return fmt.Errorf("Creating new child Thing failed: %s", err)
		}
		b.children[msg.Id] = child
	} else {
		if child.model != msg.Model {
			return fmt.Errorf("Model mismatch")
		}
		if child.name != msg.Name {
			return fmt.Errorf("Name mismatch")
		}
	}

	child.startupTime = msg.StartupTime

	b.runChild(p, child)

	return nil
}

type msgChild struct {
	Msg    string
	Id     string
	Model  string
	Name   string
	Status string
}

type msgChildren struct {
	Msg      string
	Children []msgChild
}

func (b *bridge) getOnlyChild(p *Packet) {
	for _, child := range b.children {
		resp := msgChild{
			Msg:    "ReplyOnlyChild",
			Id:     child.id,
			Model:  child.model,
			Name:   child.name,
			Status: child.status,
		}
		p.Marshal(&resp).Reply()
		break
	}
}

func (b *bridge) getChildren(p *Packet) {
	resp := msgChildren{Msg: "ReplyChildren"}
	for _, child := range b.children {
		resp.Children = append(resp.Children, msgChild{"", child.id,
			child.model, child.name, child.status})
	}
	p.Marshal(&resp).Reply()
}

func (b *bridge) getPort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	port := b.ports.getPort(id)

	switch port {
	case -1:
		fmt.Fprintf(w, "no ports available")
	case -2:
		fmt.Fprintf(w, "port busy")
	default:
		fmt.Fprintf(w, "%d", port)
	}
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
