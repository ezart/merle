package main

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/thermo"
	"log"
)

func main() {
	thing := merle.NewThing(thermo.NewThermo())

	thing.Cfg.Model = "thermo"
	thing.Cfg.Name = "thermy"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPublicTLS = 443
	thing.Cfg.PortPrivate = 8080

	log.Fatalln(thing.Run())
}
