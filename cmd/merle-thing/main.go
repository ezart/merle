// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"github.com/scottfeldman/merle/factory"
	"log"
	"os"
)

func main() {

	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	t := factory.NewThing(cfg.Thing.Id, cfg.Thing.Model, cfg.Thing.Name)
	if t == nil {
		log.Fatalf("No model named '%s'", cfg.Thing.Model)
	}

	t.Start()

	/*
	d := merle.NewDevice(m, false, cfg.Device.Id, cfg.Device.Model,
		cfg.Device.Name, "online", time.Now())
	if d == nil {
		log.Fatalf("Device creation failed, model '%s'", cfg.Device.Model)
	}

	err := d.Run(cfg.Device.User, cfg.Device.PortPublic,
		cfg.Device.PortPrivate, cfg.Hub.Host, cfg.Hub.User,
		cfg.Hub.Key)
	log.Fatalln(err)
	*/
}
