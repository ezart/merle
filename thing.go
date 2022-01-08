package merle

import (
	"fmt"
	"log"
	"time"
	"html/template"
	"net"
)

type Children    map[string]*Thing
type Subscribers map[string][]func(*Packet)

type IThing interface {
	Subscribe() Subscribers
	Config(Configurator) error
	Template() string
	Run(p *Packet)
}

type Thing struct {
	thing       IThing
	children    Children
	status      string
	id          string
	model       string
	name        string
	config      Configurator
	startupTime time.Time
	bus         *bus
	private     IWeb
	public      IWeb
	templ       *template.Template
	templErr    error
}

func defaultId(id string) string {
	if id == "" {
		// Use the MAC address of the first non-lo interface
		ifaces, err := net.Interfaces()
		if err == nil {
			for _, iface := range ifaces {
				if iface.Name != "lo" {
					id = iface.HardwareAddr.String()
					log.Println("Defaulting ID to", id)
					break
				}
			}
		}
	}
	return id
}

func must(err error) error {
	if err != nil {
		log.Println(err)
	}
	return err
}

func NewThing(thing IThing, config Configurator) (*Thing, error) {
	var cfg ThingConfig

	if err := must(config.Parse(&cfg)); err != nil {
		return nil, err
	}

	t := &Thing{
		thing:       thing,
		children:    make(Children),
		status:      "online",
		id:          defaultId(cfg.Thing.Id),
		model:       cfg.Thing.Model,
		name:        cfg.Thing.Name,
		config:      config,
		startupTime: time.Now(),
		bus:         NewBus(10, thing.Subscribe()),
	}

	t.private = WebPrivate(t, cfg.Thing.PortPrivate)
	t.public = WebPublic(t, cfg.Thing.User, cfg.Thing.PortPublic)
	t.templ, t.templErr = template.ParseFiles(thing.Template())

	t.bus.subscribe("GetIdentity", t.getIdentity)

	return t, nil
}

type msgIdentity struct {
	Msg         string
	Status      string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}

func (t *Thing) getIdentity(p *Packet) {
	resp := msgIdentity{
		Msg:         "ReplyIdentity",
		Status:      t.status,
		Id:          t.id,
		Model:       t.model,
		Name:        t.name,
		StartupTime: t.startupTime,
	}
	p.Marshal(&resp).Reply()
}

func (t *Thing) Id() string {
	return t.id
}

func (t *Thing) getChild(id string) *Thing {
	if child, ok := t.children[id]; ok {
		return child
	}
	return nil
}

func (t *Thing) startTunnel() {
}

func (t *Thing) stopTunnel() {
}

func (t *Thing) Start() error {
	t.private.Start()
	t.public.Start()
	t.startTunnel()

	t.thing.Run(newPacket(t.bus, nil, nil))

	t.private.Stop()
	t.public.Stop()
	t.stopTunnel()
	t.bus.close()

	return fmt.Errorf("Run() didn't run forever")
}
