package main

import (
	"flag"
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/examples/blink"
	"log"
	"os"
)

func main() {
	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	demo := flag.Bool("demo", false, "Run in demo mode; will simulate I/O")
	flag.Parse()

	blinker := blink.NewBlinker(*demo)

	thing := merle.NewThing(blinker, "", "blink", "blinky")

	thing.EnablePublicHTTP(80, /*443*/ 0, "admin", "examples/blink/assets")
	thing.SetTemplate("examples/blink/assets/templates/blink.html")

	log.Fatalln(thing.Run())
}
