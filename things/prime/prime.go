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
		{".*", nil}, // drop everything
	}
}

func (t *prime) run(p *merle.Packet) {
	select {}
}

func (t *prime) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"CmdRun", t.run},
	}
}

func (t *prime) Config(config merle.Configurator) error {
	return nil
}

func (t *prime) Template() string {
	return "web/templates/prime.html"
}
