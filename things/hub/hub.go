package hub

import (
	"log"
	"github.com/scottfeldman/merle"
)

type hub struct {
}

func NewModel(demo bool) merle.Thinger {
	return &hub{}
}

func (h *hub) BridgeSubscribe() merle.Subscribers {
	return merle.Subscribers{}
}

func (h *hub) Subscribe() merle.Subscribers {
	return merle.Subscribers{}
}

func (h *hub) Config(config merle.Configurator) error {
	return nil
}

func (h *hub) Template() string {
	return "web/templates/hub.html"
}

func (h *hub) Run(p *merle.Packet) {
	log.Println("run")
	for {}
}
