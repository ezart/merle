package merle

import (
	"fmt"
	"time"
	"html/template"
)

type Children    map[string]*Thing
type Subscribers map[string][]func(*Packet)

type Thinger interface {
	Subscribe() Subscribers
	Config(Configurator) error
	Template() string
	Run(p *Packet)
}

type Thing struct {
	thinger     Thinger
	status      string
	id          string
	model       string
	name        string
	startupTime time.Time
	config      Configurator
	bus         *bus
	tunnel      *tunnel
	private     Weber
	public      Weber
	templ       *template.Template
	templErr    error
	isBridge    bool
	bridge      *bridge
}

func NewThing(thinger Thinger, config Configurator) (*Thing, error) {
	var cfg ThingConfig
	var bridger Bridger
	var err error

	if err = must(config.Parse(&cfg)); err != nil {
		return nil, err
	}

	t := &Thing{
		thinger:     thinger,
		status:      "online",
		id:          defaultId(cfg.Thing.Id),
		model:       cfg.Thing.Model,
		name:        cfg.Thing.Name,
		startupTime: time.Now(),
		config:      config,
		bus:         NewBus(10, thinger.Subscribe()),
	}

	t.tunnel = NewTunnel(t.id, cfg.Mother.Host, cfg.Mother.User,
		cfg.Mother.Key, cfg.Mother.PortPrivate)

	t.private = WebPrivate(t, cfg.Thing.PortPrivate)
	t.public = WebPublic(t, cfg.Thing.User, cfg.Thing.PortPublic)

	t.templ, t.templErr = template.ParseFiles(thinger.Template())

	bridger, t.isBridge = t.thinger.(Bridger)
	if t.isBridge {
		t.bridge, err = newBridge(bridger, config)
		if must(err) != nil {
			return nil, err
		}
		t.private.HandleFunc("/port/{id}", t.bridge.getPort)
	}

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
	if !t.isBridge {
		return nil
	}
	if child, ok := t.bridge.children[id]; ok {
		return child
	}
	return nil
}

func (t *Thing) Start() error {

	t.private.Start()
	t.public.Start()
	t.tunnel.Start()

	if t.isBridge {
		t.bridge.Start()
	}

	t.thinger.Run(newPacket(t.bus, nil, nil))

	if t.isBridge {
		t.bridge.Stop()
	}

	t.tunnel.Stop()
	t.public.Stop()
	t.private.Stop()

	t.bus.close()

	return fmt.Errorf("Run() didn't run forever")
}
