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
	merle.Bridge
}

func (h *hub) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, h.HomeParams(r, cfg.Hub.Max))
}

func NewHub(id, model, name string) (*merle.Thing, error) {
	h := &hub{}

	h.Home = h.home

	if err := config.Parse(&cfg); err != nil {
		return nil, err
	}

	return h.InitBridge(id, model, name, cfg.Hub.Max, cfg.Hub.Match)
}
