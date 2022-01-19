package chat_bridge

import (
	"github.com/scottfeldman/merle"
	"log"
)

type thing struct {
}

func NewModel(log *log.Logger, demo bool) merle.Thinger {
	return &thing{}
}

func (t *thing) BridgeSubscribe() merle.Subscribers {
	return merle.Subscribers{
		{"SpamStatus", nil},
		{".*", merle.Broadcast},
	}
}

func (t *thing) run(p *merle.Packet) {
	select {}
}

func (t *thing) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"CmdRun", t.run},
	}
}

func (t *thing) Config(config merle.Configurator) error {
	return nil
}

func (t *thing) Template() string {
	return ""
}