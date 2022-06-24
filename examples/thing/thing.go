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
		HtmlTemplateText: "Hello!\n",
	}
}

func main() {
	thing := merle.NewThing(&thing{})
	thing.Cfg.PortPublic = 80
	thing.Run()
}
