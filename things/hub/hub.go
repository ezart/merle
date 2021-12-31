// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Hub is a Thing that connects other Things, including other Hubs.  Hub keeps
// track of Thing connection status and displays each Thing.
package hub

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/config"
	"html/template"
	"net/http"
)

var templ *template.Template

func init() {
	templ = template.Must(template.ParseFiles("web/templates/hub.html"))
}

var cfg struct {
	Hub struct {
		Max   uint   `yaml:"Max"`
		Match string `yaml:"Match"`
	} `yaml:"Hub"`
}

type hub struct {
	merle.Thing
}

func (h *hub) init() error {
	err := config.ParseFile(h.ConfigFile(), &cfg)
	if err != nil {
		return err
	}

	return h.ListenForThings(cfg.Hub.Max, cfg.Hub.Match)
}

func (h *hub) run() {
	for {}
}

func (h *hub) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, h.HomeParams(r, cfg.Hub.Max))
}

func NewThing(id, model, name string) *merle.Thing {
	h := &hub{}

	h.Init = h.init
	h.Run = h.run
	h.Home = h.home

	return h.InitThing(id, model, name)
}
