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

func (h *hello) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		HtmlTemplateText: "Hello, world!\n",
	}
}

func main() {
	thing := merle.NewThing(&hello{})
	thing.Cfg.PortPublic = 80
	log.Fatalln(thing.Run())
}
