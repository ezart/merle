// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package bridge

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/config"
	"html/template"
	"net/http"
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
	online map[string]*merle.Thing
}

func (b *bridge) init() error {
	err := config.ParseFile(b.ConfigFile(), &cfg)
	if err != nil {
		return err
	}

	b.online = make(map[string]*merle.Thing)

	return b.ListenForThings(cfg.Bridge.Max, cfg.Bridge.Match)
}

func (b *bridge) run() {
	for {
	}
}

func (b *bridge) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, b.HomeParams(r, nil))
}

func (b *bridge) connect(child *merle.Thing) {
	if child.Status() == "online" {
		b.online[child.Id()] = child
	} else {
		delete(b.online, child.Id())
	}
}

func NewThing(id, model, name string) *merle.Thing {
	b := &bridge{}

	b.Init = b.init
	b.Run = b.run
	b.Home = b.home
	b.Connect = b.connect

	return b.InitThing(id, model, name)
}
