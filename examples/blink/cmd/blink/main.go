package main

import (
	"flag"
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/examples/blink"
	"log"
	"os"
)

func flagThingConfig(id, model, name, user, assetsDir string) *merle.ThingConfig {
	var cfg merle.ThingConfig

	flag.BoolVar(&cfg.Thing.Prime, "prime", false, "Run as Thing-prime")
	flag.UintVar(&cfg.Thing.PortPrime, "pport", 0, "Prime Port")

	flag.StringVar(&cfg.Thing.Id, "id", id, "Thing ID")
	flag.StringVar(&cfg.Thing.Model, "model", model, "Thing model")
	flag.StringVar(&cfg.Thing.Name, "name", name, "Thing name")

	flag.StringVar(&cfg.Thing.User, "luser", user,
		"Local user for HTTP Basic Authentication")
	flag.UintVar(&cfg.Thing.PortPublic, "lport", 80,
		"Local public HTTP listening port")
	flag.UintVar(&cfg.Thing.PortPublicTLS, "lportTLS", 0,
		"Local public HTTPS listening port (default 0, but usually 443)")
	flag.UintVar(&cfg.Thing.PortPrivate, "lportPriv", 8080,
		"Local private HTTP listening port")
	flag.StringVar(&cfg.Thing.AssetsDir, "lassets", assetsDir,
		"Local path to assets directory")

	flag.StringVar(&cfg.Mother.Host, "rhost", "",
		"Remote host name or IP address")
	flag.StringVar(&cfg.Mother.User, "ruser", user,
		"Remote user")
	flag.StringVar(&cfg.Mother.Key, "rkey",
		"/home/" + user + "/.ssh/id_rsa", "Remote SSH identity key")
	flag.UintVar(&cfg.Mother.PortPrivate, "rport", 8080,
		"Remote private HTTP listening port")

	return &cfg
}

func main() {
	var thing *merle.Thing

	demo := flag.Bool("demo", false, "Run in demo mode; will simulate I/O")
	cfg := flagThingConfig("", "blink", "blinky", "admin",
		"examples/blink/assets")
	flag.Parse()

	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	blinker := blink.NewBlinker(*demo)
	thing = merle.NewThing(blinker, cfg)

	log.Fatalln(thing.Run())
}
