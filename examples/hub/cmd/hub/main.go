package main

import (
	"log"

	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/hub"
)

func main() {
	thing := merle.NewThing(hub.NewHub())

	thing.Cfg.Model = "hub"
	thing.Cfg.Name = "hubby"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 6000
	thing.Cfg.PortPublicTLS = 443

	log.Fatalln(thing.Run())
}
