package chat

import (
	"github.com/scottfeldman/merle"
	"log"
)

type chat struct {
}

func NewModel(log *log.Logger, demo bool) merle.Thinger {
	return &chat{}
}

func (c *chat) run(p *merle.Packet) {
	select{}
}

func (c *chat) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"CmdRun", c.run},
		{"CmdNewUser", merle.Broadcast},
		{"CmdText", merle.Broadcast},
		{"CmdStart", nil},
	}
}

func (c *chat) Config(config merle.Configurator) error {
	return nil
}

func (c *chat) Template() string {
	return "web/templates/chat.html"
}
