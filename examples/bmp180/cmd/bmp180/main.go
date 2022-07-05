package main

import (
	"flag"
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/bmp180"
	"log"
)

func main() {
	bmp180 := bmp180.NewBmp180()
	thing := merle.NewThing(bmp180)

	thing.Cfg.Model = "bmp180"
	thing.Cfg.Name = "bumpy"
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
