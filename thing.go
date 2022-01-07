package merle

import (
	"log"
	"time"
)

type Subscribers map[string]func()

type IThing interface {
	Subscribe() Subscribers
	Config(Configurator) error
	Run() error
}

type Thing struct {
	thing IThing
	status string
	id	string
	model	string
	name	string
	config	Configurator
	subs	Subscribers
	demo	bool
	startupTime time.Time
	bus *bus
}

func defaultId(id string) string {
	if id == "" {
		id = "012345"
	}
	return id
}

func must(err error) error {
	if err != nil {
		log.Println(err)
	}
	return err
}

func NewThing(_thing IThing, _config Configurator, _demo bool) (*Thing, error) {
	var cfg ThingConfig

	if err := must(_config.Parse(&cfg)); err != nil {
		return nil, err
	}

	return &Thing{
		thing: _thing,
		status: "online",
		id: defaultId(cfg.Thing.Id),
		model: cfg.Thing.Model,
		name: cfg.Thing.Name,
		config: _config,
		subs: make(Subscribers),
		demo: _demo,
		startupTime: time.Now(),
		bus: NewBus(10),
	}, nil
}

func (t *Thing) StartHttp(cfg *ThingConfig) error {
	return nil
}

func (t *Thing) StartTunnel(cfg *ThingConfig) error {
	return nil
}

func (t *Thing) Start() error {
	var cfg ThingConfig

	if err := must(t.config.Parse(&cfg)); err != nil {
		return err
	}

	if err := must(t.StartHttp(&cfg)); err != nil {
		return err
	}

	if err := must(t.StartTunnel(&cfg)); err != nil {
		return err
	}

	if err := must(t.thing.Run()); err != nil {
		return err
	}

	return nil
}
