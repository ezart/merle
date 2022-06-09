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
		TemplateText: "Hello, world!\n",
	}
}

func main() {
	thing := merle.NewThing(&hello{})
	log.Fatalln(thing.Run())
}
