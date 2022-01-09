package hub

import (
	"github.com/scottfeldman/merle"
	"log"
)

type hub struct {
	log *log.Logger
}

func NewModel(l *log.Logger, demo bool) merle.Thinger {
	return &hub{log: l}
}

func (h *hub) all(p *merle.Packet) {
	h.log.Println("HUB Receive:", p.String())
}

func (h *hub) BridgeSubscribe() merle.Subscribers {
	return merle.Subscribers{
		".*": h.all,
		"-CmdPause": nil,
	}
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
	for {}
}
