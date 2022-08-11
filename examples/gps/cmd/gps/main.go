package main

import (
	"flag"
	"log"

	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/gps"
)

func main() {
	gps := gps.NewGps()
	thing := merle.NewThing(gps)

	thing.Cfg.Model = "gps"
	thing.Cfg.Name = "gypsy"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 6000

	flag.BoolVar(&gps.Demo, "demo", false, "Run in Demo mode")

	flag.StringVar(&thing.Cfg.MotherHost, "rhost", "", "Remote host")
	flag.StringVar(&thing.Cfg.MotherUser, "ruser", "merle", "Remote user")
	flag.BoolVar(&thing.Cfg.IsPrime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")

	flag.Parse()

	log.Fatalln(thing.Run())
}
