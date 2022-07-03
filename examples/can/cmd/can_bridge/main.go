package main

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/can"
	"log"
)

func main() {
	bridge := can.NewBridge()
	thing := merle.NewThing(bridge)

	thing.Cfg.Model = "bridge"
	thing.Cfg.Name = "bridgy"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 8080

	log.Fatalln(thing.Run())
}
