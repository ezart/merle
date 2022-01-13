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

func (t *hello_world) Subscribe() merle.Subscribers {
	return merle.Subscribers{}
}

func (t *hello_world) Config(config merle.Configurator) error {
	return nil
}

func (t *hello_world) Template() string {
	return "web/templates/hello_world.html"
}

func (t *hello_world) Run(p *merle.Packet) {
	t.log.Println("Hello World!")
	select{}
}
