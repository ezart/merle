// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/stork"
	"log"
	"os"
)

func main() {
	if os.Geteuid() != 0 {
		log.Fatalln("Must run as root")
	}

	log.SetFlags(0)

	cfgFile := flag.String("config", "/etc/merle/thing.yml", "Config File")
	demo := flag.Bool("demo", false, "Run Thing in demo mode; will simulate I/O")
	models := flag.Bool("models", false, "Print a list models")

	flag.Parse()

	stork := stork.NewStork()

	if *models {
		for _, model := range stork.Models() {
			fmt.Println(model)
		}
		return
	}

	config := merle.NewYamlConfig(*cfgFile)

	log.Fatalln(merle.RunThing(stork, config, *demo))
}
