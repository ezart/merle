package main

import (
	"flag"
	"log"

	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/relays"
)

func main() {
	thing := merle.NewThing(relays.NewRelays())

	thing.Cfg.Model = "relays"
	thing.Cfg.Name = "relaysforhope"
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
