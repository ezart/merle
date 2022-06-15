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

func main() {
	thing := merle.NewThing(&thing{})
	thing.Cfg.PortPublic = 80
	thing.Cfg.HtmlTemplateText = "Hello!\n"
	thing.Run()
}
