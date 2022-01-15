package hello_world

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

func (t *thing) run(p *merle.Packet) {
	t.log.Println("Hello World!")
	select{}
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
	return "web/templates/hello_world.html"
}
