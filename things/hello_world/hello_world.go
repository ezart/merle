package hello_world

import (
	"github.com/scottfeldman/merle"
	"log"
)

type hello_world struct {
	log *log.Logger
}

func NewModel(log *log.Logger, demo bool) merle.Thinger {
	return &hello_world{log: log}
}

func (t *hello_world) run(p *merle.Packet) {
	t.log.Println("Hello World!")
	select{}
}

func (t *hello_world) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"CmdRun", t.run},
	}
}

func (t *hello_world) Config(config merle.Configurator) error {
	return nil
}

func (t *hello_world) Template() string {
	return "web/templates/hello_world.html"
}
