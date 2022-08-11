package main

import (
	"flag"
	"log"

	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/thermo"
)

func main() {
	thing := merle.NewThing(thermo.NewThermo())

	thing.Cfg.Model = "thermo"
	thing.Cfg.Name = "thermy"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPrivate = 8080

	flag.StringVar(&thing.Cfg.MotherHost, "rhost", "", "Remote host")
	flag.StringVar(&thing.Cfg.MotherUser, "ruser", "merle", "Remote user")
	flag.BoolVar(&thing.Cfg.IsPrime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")

	flag.Parse()

	log.Fatalln(thing.Run())
}
