package main

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/hub"
	"log"
)

func main() {
	hub := hub.NewHub()
	thing := merle.NewThing(hub)

	thing.Cfg.Model = "hub"
	thing.Cfg.Name = "hubby"
	thing.Cfg.User = "merle"

	log.Fatalln(thing.Run())
}
