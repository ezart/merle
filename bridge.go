package merle

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type bridgeConfig struct {
	Bridge struct {
		Max   uint   `yaml:"Max"`
		Match string `yaml:"Match"`
	} `yaml:"Bridge"`
}

type bridger interface {
	BridgeSubscribe() Subscribers
}

type children map[string]*Thing

type bridge struct {
	stork    Storker
	thing    *Thing
	children children
	bus      *bus
	ports    *ports
}

func newBridge(stork Storker, config Configurator, thing *Thing) (*bridge, error) {
	var cfg bridgeConfig

	if err := must(config.Parse(&cfg)); err != nil {
		return nil, err
	}

	bridger := thing.thinger.(bridger)

	b := &bridge{
		stork:    stork,
		thing:    thing,
		children: make(children),
		bus:      newBus(thing.log, 10, bridger.BridgeSubscribe()),
	}

	b.ports = newPorts(thing.log, cfg.Bridge.Max, cfg.Bridge.Match, b.attachCb)

	b.thing.bus.subscribe("GetChildren", b.getChildren)
	b.thing.private.HandleFunc("/port/{id}", b.getPort)

	return b, nil
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

func (b *bridge) changeStatus(child *Thing, status string) {
	child.status = status

	spam := SpamStatus{
		Msg:    "SpamStatus",
		Id:     child.id,
		Model:  child.model,
		Name:   child.name,
		Status: child.status,
	}
	newPacket(b.thing.bus, nil, &spam).Broadcast()
	b.bus.receive(newPacket(b.bus, nil, &spam))
}

func (b *bridge) attachCb(p *port, msg *msgIdentity) error {
	var err error

	// TODO think about if it makes sense to allow you to be your own Mother?

	if b.thing.Id() == msg.Id {
		return fmt.Errorf("Sorry, you can't be your own Mother")
	}

	child := b.getChild(msg.Id)

	if child == nil {
		config := newChildConfig(msg.Id, msg.Model, msg.Name)
		child, err = NewThing(b.stork, config, false)
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

	b.changeStatus(child, "online")
	child.runInBridge(p)
	b.changeStatus(child, "offline")

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
	resp := msgChildren{ Msg: "ReplyChildren" }
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
	must(b.ports.Start())
}

func (b *bridge) Stop() {
	b.ports.Stop()
	b.bus.close()
}
