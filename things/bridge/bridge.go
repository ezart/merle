// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bridge

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/config"
	"html/template"
	"net/http"
	"log"
)

var templ *template.Template

func init() {
	templ = template.Must(template.ParseFiles("web/templates/bridge.html"))
}

var cfg struct {
	Bridge struct {
		Max   uint   `yaml:"Max"`
		Match string `yaml:"Match"`
	} `yaml:"Bridge"`
}

type bridge struct {
	merle.Thing
	// TODO Need R/W lock for map[]
	online map[*merle.Thing]bool
}

func (b *bridge) bridgePkt(child *merle.Thing, p *merle.Packet) {
	for thing, _ := range b.online {
		if thing != child {
		}
	}
}

func (b *bridge) connect(child *merle.Thing) {
	if child.Status() == "online" {
		b.ChildSubscribe(child, ".*", bridgePkt)
		b.online[child] = true
	} else {
		b.ChildUnSubscribe(child, ".*", bridgePkt)
		delete(b.online, child)
	}
}

func (b *bridge) init(soft bool) error {
	err := config.ParseFile(b.ConfigFile(), &cfg)
	if err != nil {
		return err
	}

	return b.ListenForChildren(cfg.Bridge.Max, cfg.Bridge.Match, b.connect)
}

func (b *bridge) run() {
	for {
	}
}

func (b *bridge) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, b.HomeParams(r, nil))
}

func NewBridge(id, model, name string) *merle.Thing {
	b := &bridge{}

	b.Init = b.init
	b.Run = b.run
	b.Home = b.home

	b.online = make(map[*merle.Thing]bool)

	return b.InitThing(id, model, name)
}
