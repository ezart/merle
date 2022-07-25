// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package hub

import (
	"testing"

	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/gps"
	"github.com/merliot/merle/examples/relays"
)

func testHub(t *testing.T, hub *merle.Thing) {
	// sleep a second for http servers to start
	/*
		time.Sleep(time.Second)
		testHomePage(t, publicPort)
		testIdentify(t, thing, privatePort)
		testDone(t, thing, privatePort)
	*/
}

func TestRun(t *testing.T) {
	newGps := gps.NewGps()
	gps := merle.NewThing(newGps)
	if gps == nil {
		t.Errorf("Create new gps Thing failed")
	}

	newGps.Demo = true
	gps.Cfg.Id = "gps01"
	gps.Cfg.Model = "gps"
	gps.Cfg.Name = "gypsy"
	gps.Cfg.PortPrivate = 6000
	gps.Cfg.MotherHost = "localhost"
	gps.Cfg.MotherUser = "merle"

	go gps.Run()

	relays := merle.NewThing(relays.NewRelays())
	if relays == nil {
		t.Errorf("Create new relays Thing failed")
	}

	relays.Cfg.Id = "relays01"
	relays.Cfg.Model = "relays"
	relays.Cfg.Name = "nothere"
	relays.Cfg.PortPrivate = 6001
	relays.Cfg.MotherHost = "localhost"
	relays.Cfg.MotherUser = "merle"

	go relays.Run()

	hub := merle.NewThing(NewHub())
	if hub == nil {
		t.Errorf("Create new hub Thing failed")
	}

	hub.Cfg.Model = "hub_test"
	hub.Cfg.Name = "hubby_test"

	hub.Cfg.PortPublic = 7000
	hub.Cfg.PortPrivate = 8080

	go testHub(t, hub)

	err := hub.Run()
	if err == nil {
		t.Errorf("Run should have errored out")
	}
}
