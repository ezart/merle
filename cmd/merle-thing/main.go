// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"flag"
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/stork"
	"log"
	"os"
)

func must(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	var cfg merle.ThingConfig

	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	log.SetFlags(0)

	cfgFile := flag.String("config", "/etc/merle/thing.yml", "Config File")
	demo := flag.Bool("demo", false, "Run Thing in demo mode; will simulate I/O")

	flag.Parse()

	config := merle.NewYamlConfig(*cfgFile)
	must(config.Parse(&cfg))
	log.Printf("%+v", cfg)

	model, err := stork.NewModel(cfg.Thing.Model, *demo)
	must(err)

	thing, err := merle.NewThing(model, config)
	must(err)

	must(thing.Start())
}
