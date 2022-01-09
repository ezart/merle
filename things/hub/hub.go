package hub

import (
	"github.com/scottfeldman/merle"
	"log"
)

type hub struct {
}

func NewModel(demo bool) merle.Thinger {
	return &hub{}
}

func (h *hub) spamStatus(p *merle.Packet) {
	var spam merle.SpamStatus
	p.Unmarshal(&spam)
	log.Println("SPAM", spam)
}

func (h *hub) BridgeSubscribe() merle.Subscribers {
	return merle.Subscribers{
		"SpamStatus":    {h.spamStatus},
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
