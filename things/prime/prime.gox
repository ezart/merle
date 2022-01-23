package prime

import (
	"github.com/scottfeldman/merle"
	"log"
)

type thing struct {
	log *log.Logger
}

func NewModel(log *log.Logger, demo bool) merle.Thinger {
	return &thing{log: log}
}

func (t *thing) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		// drop everything
	}
}

func (t *thing) run(p *merle.Packet) {
	select {}
}

func (t *thing) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		"CmdRun": t.run,
	}
}

func (t *thing) Config(config merle.Configurator) error {
	return nil
}

func (t *thing) Template() string {
	return "web/templates/prime.html"
}
