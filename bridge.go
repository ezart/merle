package merle

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// Bridge configuration
type bridgeConfig struct {
	Bridge struct {
		// Maximum number of things (children) that can connect to the
		// bridge
		Max   uint   `yaml:"Max"`
		// Match is a regular expresion (re) to specifiy which things
		// can connect to the bridge.  The re matches against three
		// fields of the thing, ID, Model, and Name.  The re is
		// composed with these three fields seperated by ":" character:
		// "ID:Model:Name".  See
		// https://github.com/google/re2/wiki/Syntax for regular
		// expression syntax.  Examples:
		//
		//	".*:.*:.*"		Match any thing.
		//	"123456:.*:.*"		Match only a thing with ID=123456
		//	".*:chat:.*"		Match only chat things
		Match string `yaml:"Match"`
	} `yaml:"Bridge"`
}

// A thing implementing the bridger interface is a bridge
type bridger interface {
	// List of subscribers on bridge bus.  All packets from all connected
	// things (children) are forwarded to the bridge bus and tested against
	// these subscribers.  To ignore all packets on the bridge bus, install
	// the subscriber {".*", nil}.  This will drop all packets.
	BridgeSubscribe() Subscribers
}

// Children are the things connected to the bridge
type children map[string]*thing

type bridge struct {
	log      *log.Logger
	stork    Storker
	thing    *thing
	children children
	bus      *bus
	ports    *ports
}

func newBridge(log *log.Logger, stork Storker, config Configurator, thing *thing) (*bridge, error) {
	var cfg bridgeConfig

	if err := config.Parse(&cfg); err != nil {
		log.Println("Configure bridge error:", err)
		return nil, err
	}

	bridger := thing.thinger.(bridger)

	b := &bridge{
		log:      log,
		stork:    stork,
		thing:    thing,
		children: make(children),
		bus:      newBus(thing.log, 10, bridger.BridgeSubscribe()),
	}

	b.ports = newPorts(thing.log, cfg.Bridge.Max, cfg.Bridge.Match, b.attachCb)

	b.thing.bus.subscribe("GetChildren", b.getChildren)
	b.thing.private.handleFunc("/port/{id}", b.getPort)

	return b, nil
}

func (b *bridge) getChild(id string) *thing {
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

func (b *bridge) changeStatus(child *thing, sock *wireSocket, status string) {
	child.status = status

	spam := SpamStatus{
		Msg:    "SpamStatus",
		Id:     child.id,
		Model:  child.model,
		Name:   child.name,
		Status: child.status,
	}
	newPacket(b.thing.bus, nil, &spam).Broadcast()
	b.bus.receive(newPacket(b.bus, sock, &spam))
}

func (b *bridge) runChild(p *port, child *thing) {
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

func (b *bridge) attachCb(p *port, msg *msgIdentity) error {
	var err error

	// TODO think about if it makes sense to allow you to be your own Mother?

	if b.thing.id == msg.Id {
		return fmt.Errorf("Sorry, you can't be your own Mother")
	}

	child := b.getChild(msg.Id)

	if child == nil {
		config := newChildConfig(msg.Id, msg.Model, msg.Name)
		child, err = newThing(b.stork, config, false)
		if err != nil {
			return fmt.Errorf("Creating new Thing failed: %s", err)
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
	Id     string
	Model  string
	Name   string
	Status string
}

type msgChildren struct {
	Msg      string
	Children []msgChild
}

func (b *bridge) getChildren(p *Packet) {
	resp := msgChildren{Msg: "ReplyChildren"}
	for _, child := range b.children {
		resp.Children = append(resp.Children, msgChild{child.id,
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

func (b *bridge) Start() {
	if err := b.ports.Start(); err != nil {
		b.log.Println("Starting bridge error:", err)
	}
}

func (b *bridge) Stop() {
	b.ports.Stop()
	b.bus.close()
}
