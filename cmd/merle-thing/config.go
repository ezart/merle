// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type config struct {
	Thing struct {
		Id    string `yaml:"Id"`
		Model string `yaml:"Model"`
		Name  string `yaml:"Name"`
		User  string `yaml:"User"`
		PortPublic int `yaml:"PortPublic"`
		PortPrivate int `yaml:"PortPrivate"`
	} `yaml:"Thing"`
	Hub struct {
		Host string `yaml:"Host"`
		User string `yaml:"User"`
		Key  string `yaml:"Key"`
	} `yaml:"Hub"`
}

var cfg config

func init() {
	log.SetFlags(0)

	f, err := os.Open("/etc/merle/thing.yml")
	if err != nil {
		log.Fatalln("Opening config file:", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatalln("Decoding config file:", err)
	}

	log.Printf("Config: %+v", cfg)
}
