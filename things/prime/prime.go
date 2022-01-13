package prime

import (
	"github.com/scottfeldman/merle"
	"log"
)

type prime struct {
	log *log.Logger
}

func NewModel(log *log.Logger, demo bool) merle.Thinger {
	return &prime{log: log}
}

func (t *prime) BridgeSubscribe() merle.Subscribers {
	return merle.Subscribers{
		{ ".*", nil }, // drop everything
	}
}

func (t *prime) Subscribe() merle.Subscribers {
	return merle.Subscribers{}
}

func (t *prime) Config(config merle.Configurator) error {
	return nil
}

func (t *prime) Template() string {
	return "web/templates/prime.html"
}

func (t *prime) Run(p *merle.Packet) {
	for {
	}
}
