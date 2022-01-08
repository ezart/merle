package skeleton

import (
	"github.com/scottfeldman/merle"
)

type skeleton struct {
}

func NewModel(demo bool) merle.IThing {
	return &skeleton{}
}

func (t *skeleton) Subscribe() merle.Subscribers {
	return merle.Subscribers{}
}

func (t *skeleton) Config(config merle.Configurator) error {
	return nil
}

func (t *skeleton) Template() string {
	return "web/templates/skeleton.html"
}

func (t *skeleton) Run(p *merle.Packet) {
	for {}
}
