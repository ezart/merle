// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"github.com/scottfeldman/merle/factory"
	"flag"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	cfgFile  := flag.String("config", "/etc/merle/thing.yml", "Config File")
	flag.Parse()

	cfg := parseCfgFile(*cfgFile)

	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	t := factory.NewThing(cfg.Thing.Id, cfg.Thing.Model, cfg.Thing.Name)
	if t == nil {
		log.Fatalf("No model named '%s'", cfg.Thing.Model)
	}

	t.SetFactory(factory.NewThing)

	t.TunnelConfig(cfg.Mother.Host, cfg.Mother.User, cfg.Mother.Key,
		cfg.Mother.PortPrivate)

	t.HttpConfig(cfg.Thing.User, cfg.Thing.PortPublic, cfg.Thing.PortPrivate)

	t.Start()
}
