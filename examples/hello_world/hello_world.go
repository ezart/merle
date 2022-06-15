package main

import (
	"github.com/merliot/merle"
	"log"
)

type hello struct {
}

func (h *hello) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: merle.RunForever,
	}
}

func main() {
	thing := merle.NewThing(&hello{})
	thing.Cfg.HtmlTemplateText = "Hello, world!\n"
	log.Fatalln(thing.Run())
}
