package main

import (
	"github.com/scottfeldman/merle"
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
		TemplateText: "Hello, world!\n",
	}
}

func main() {
	var cfg merle.ThingConfig

	log.SetFlags(0)

	cfg.Thing.PortPublic = 80

	thing := merle.NewThing(&hello{}, &cfg)

	log.Fatalln(thing.Run())
}
