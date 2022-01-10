package skeleton

import (
	"github.com/scottfeldman/merle"
	"log"
)

type skeleton struct {
}

func NewModel(l *log.Logger, demo bool) merle.Thinger {
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
	for {
	}
}