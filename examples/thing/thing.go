package main

import (
	"github.com/merliot/merle"
)

type thing struct {
}

func (t *thing) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: merle.RunForever,
	}
}

func (t *thing) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		TemplateText: "Hello!\n",
	}
}

func main() {
	thing := merle.NewThing(&thing{})
	thing.Run()
}
