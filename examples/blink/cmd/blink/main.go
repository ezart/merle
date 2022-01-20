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

	prime := flag.Bool("prime", false, "Run as Thing-prime")
	demo := flag.Bool("demo", false, "Run in demo mode; will simulate I/O")

	id := flag.String("id", "", "Thing ID")
	model := flag.String("model", "blink", "Thing model")
	name := flag.String("name", "blinky", "Thing name")

	port := flag.Uint("port", 80, "Local public HTTP listening port")
	portTLS := flag.Uint("portTLS", 443, "Local public HTTPS listening port")
	portPriv := flag.Uint("portPriv", 8080, "Local private HTTP listening port")
	user := flag.String("user", "admin", "Local user for HTTP Basic Authentication")

	primeHost := flag.String("rhost", "", "Remote host name or IP address")
	primeUser := flag.String("ruser", "admin", "Remote user")
	primeKey := flag.String("rkey", "/home/admin/.ssh/id_rsa", "Remote SSH identity key")
	primePortPriv := flag.Uint("rport", 8080, "Remote private HTTP listening port")

	flag.Parse()

	blinker := blink.NewBlinker(*demo)

	thing := merle.NewThing(blinker, *id, *model, *name)

	thing.EnablePublicHTTP(*port, *portTLS, *user, "examples/blink/assets")
	thing.EnablePrivateHTTP(*portPriv)
	thing.EnableTunnel(*primeHost, *primeUser, *primeKey, *portPriv, *primePortPriv)

	thing.SetTemplate("examples/blink/assets/templates/blink.html")

	if *prime {
		log.Fatalln(thing.RunPrime())
	} else {
		log.Fatalln(thing.Run())
	}
}
