package main

import (
	"flag"
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/blink"
	"log"
)

func main() {
	demo := flag.Bool("demo", false, "Run in demo mode; will simulate I/O")
	flag.Parse()

	blinker := blink.NewBlinker(*demo)
	thing := merle.NewThing(blinker)

	thing.Cfg.Model = "blink"
	thing.Cfg.Name = "blinky"
	thing.Cfg.User = "merle"

	log.Fatalln(thing.Run())
}
