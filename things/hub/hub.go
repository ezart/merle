// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Hub is a Thing that connects other Things (children).  Hub keeps track of
// Child connection status and displays each Child.
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

func (h *hub) init(soft bool) error {
	err := config.ParseFile(h.ConfigFile(), &cfg)
	if err != nil {
		return err
	}

	return h.ListenForChildren(cfg.Hub.Max, cfg.Hub.Match, nil)
}

func (h *hub) run() {
	for {
	}
}

func (h *hub) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, h.HomeParams(r, cfg.Hub.Max))
}

func NewHub(id, model, name string) *merle.Thing {
	h := &hub{}

	h.Init = h.init
	h.Run = h.run
	h.Home = h.home

	return h.InitThing(id, model, name)
}
