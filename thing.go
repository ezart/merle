package merle

import (
	"log"
)

type Subscribers map[string]func()

type IThing interface {
	Subscribe() Subscribers
	Config(Configurator) error
}

type Thing struct {
	thing	IThing
	id	string
	model	string
	name	string
	config	Configurator
	subs	Subscribers
	demo	bool
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
		id: defaultId(cfg.Thing.Id),
		model: cfg.Thing.Model,
		name: cfg.Thing.Name,
		config: _config,
		subs: make(Subscribers),
		demo: _demo,
	}, nil
}

func (t *Thing) Http(authUser string, portPublic uint, portPrivate uint) error {
	return nil
}

func (t *Thing) Start() error {
	return nil
}
