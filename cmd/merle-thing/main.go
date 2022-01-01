// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"github.com/scottfeldman/merle/config"
	"github.com/scottfeldman/merle/stork"
	"flag"
	"log"
	"os"
)

type cfg struct {
	Thing struct {
		Id          string `yaml:"Id"`
		Model       string `yaml:"Model"`
		Name        string `yaml:"Name"`
		User        string `yaml:"User"`
		PortPublic  int    `yaml:"PortPublic"`
		PortPrivate int    `yaml:"PortPrivate"`
	} `yaml:"Thing"`
	Mother struct {
		Host string `yaml:"Host"`
		User string `yaml:"User"`
		Key  string `yaml:"Key"`
		PortPrivate int `yaml:"PortPrivate"`
	} `yaml:"Mother"`
}

func main() {
	var cfg cfg

	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	log.SetFlags(0)

	cfgFile  := flag.String("config", "/etc/merle/thing.yml", "Config File")
	demoMode := flag.Bool("demo", false, "Run Thing in demo mode; will simulate I/O")

	flag.Parse()

	err := config.ParseFile(*cfgFile, &cfg)
	if err != nil {
		log.Fatalln("Opening config file:", err)
	}

	t := stork.NewThing(cfg.Thing.Id, cfg.Thing.Model, cfg.Thing.Name)
	if t == nil {
		log.Fatalf("No model named '%s'", cfg.Thing.Model)
	}

	t.SetConfigFile(*cfgFile)
	t.SetDemoMode(*demoMode)
	t.SetStork(stork.NewThing)

	t.TunnelConfig(cfg.Mother.Host, cfg.Mother.User, cfg.Mother.Key,
		cfg.Mother.PortPrivate)

	t.HttpConfig(cfg.Thing.User, cfg.Thing.PortPublic, cfg.Thing.PortPrivate)

	t.Start()
}
