// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var cfgFile string

func SetFile(file string) {
	cfgFile = file
}

func Parse(cfg interface{}) error {
	f, err := os.Open(cfgFile)
	if err != nil {
		return fmt.Errorf("Opening config file failure: %s", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("Config decode error: %s", err)
	}

	log.Printf("Config [%s] %+v", cfgFile, cfg)

	return nil
}
