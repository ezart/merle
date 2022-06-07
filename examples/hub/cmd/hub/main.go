package main

import (
	"flag"
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/hub"
	"log"
	"os"
)

func main() {
	cfg := merle.FlagBridgeConfig("", "hub", "hubby", "admin", 8100, 8200)
	flag.Parse()

	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	hub := hub.NewHub()

	thing := merle.NewThing(hub, cfg)

	log.Fatalln(thing.Run())
}
