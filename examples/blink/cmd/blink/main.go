package main

import (
	"flag"
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/examples/blink"
	"log"
	"os"
)

func main() {
	demo := flag.Bool("demo", false, "Run in demo mode; will simulate I/O")
	cfg := merle.FlagThingConfig("", "blink", "blinky", "admin",
		"examples/blink/assets")
	flag.Parse()

	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	blinker := blink.NewBlinker(*demo)
	thing := merle.NewThing(blinker, cfg)

	log.Fatalln(thing.Run())
}
