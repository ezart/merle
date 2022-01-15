package hub

import (
	"github.com/scottfeldman/merle"
	"log"
)

type hub struct {
	log *log.Logger
}

func NewModel(log *log.Logger, demo bool) merle.Thinger {
	return &hub{log: log}
}

func (h *hub) BridgeSubscribe() merle.Subscribers {
	return merle.Subscribers{
		{".*", nil}, // drop everything
	}
}

func (h *hub) run(p *merle.Packet) {
	select {}
}

func (h *hub) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"CmdRun", h.run},
	}
}

func (h *hub) Config(config merle.Configurator) error {
	return nil
}

func (h *hub) Template() string {
	return "web/templates/hub.html"
}
