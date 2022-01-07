package merle

import (
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
}

func defaultId(id string) string {
	if id == "" {
		id = "012345"
	}
	return id
}

func NewThing(_thing IThing, _config Configurator) *Thing {
	var cfg thingConfig

	if err := _config.Parse(&cfg); err != nil {
		return nil
	}

	return &Thing{
		thing: _thing,
		id: defaultId(cfg.Thing.Id),
		model: cfg.Thing.Model,
		name: cfg.Thing.Name,
		config: _config,
	}
}

func (t *Thing) Http(authUser string, portPublic uint, portPrivate uint) error {
	return nil
}

func (t *Thing) Run() error {
	return nil
}
