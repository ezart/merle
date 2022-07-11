package main

import (
	"github.com/merliot/merle"
)

type thing struct {
}

func (t *thing) Subscribers() merle.Subscribers { ... }

func (t *thing) Assets() *merle.ThingAssets { ... }

func main() {
	thing := merle.NewThing(&thing{})
	thing.Cfg.PortPublic = 80
	thing.Run()
}
