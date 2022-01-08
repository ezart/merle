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
	bridger     Bridger
	isBridge    bool
	children    Children
	bridgeBus   *bus
	ports       *ports
	tunnel      *tunnel
	private     IWeb
	public      IWeb
	templ       *template.Template
	templErr    error
}

func NewThing(thinger Thinger, config Configurator) (*Thing, error) {
	var cfg ThingConfig

	if err := must(config.Parse(&cfg)); err != nil {
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

	t.bridger, t.isBridge = t.thinger.(Bridger)
	if t.isBridge {
		t.children = make(Children)
		t.bridgeBus = NewBus(10, t.bridger.BridgeSubscribe())
		t.ports = NewPorts(20, ".*")
	}

	t.tunnel = NewTunnel(t.id, cfg.Mother.Host, cfg.Mother.User,
		cfg.Mother.Key, cfg.Mother.PortPrivate)

	t.private = WebPrivate(t, cfg.Thing.PortPrivate)
	t.public = WebPublic(t, cfg.Thing.User, cfg.Thing.PortPublic)

	t.templ, t.templErr = template.ParseFiles(thinger.Template())

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

func (t *Thing) Start() error {

	if t.isBridge {
		t.ports.Start()
	}

	t.private.Start()
	t.public.Start()
	t.tunnel.Start()

	t.thinger.Run(newPacket(t.bus, nil, nil))

	t.tunnel.Stop()
	t.public.Stop()
	t.private.Stop()

	if t.isBridge {
		t.ports.Stop()
	}

	t.bus.close()
	t.bridgeBus.close()

	return fmt.Errorf("Run() didn't run forever")
}
