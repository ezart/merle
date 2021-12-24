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

	//t.TunnelConfig(cfg.Hub.Host, cfg.Hub.User, cfg.Hub.Key)
	t.HttpConfig(cfg.Thing.User, cfg.Thing.PortPublic, cfg.Thing.PortPrivate)

	t.Start(true)
}
