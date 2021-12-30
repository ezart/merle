// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Shadow is a Thing's shadow.
package shadow

import (
	"github.com/scottfeldman/merle"
	"html/template"
	"net/http"
)

var templ *template.Template

func init() {
	templ = template.Must(template.ParseFiles("web/templates/shadow.html"))
}

type shadow struct {
	merle.Thing
}

func (s *shadow) init() error {
	return s.ListenForThings(1, "[*][*][*]")
}

func (s *shadow) run() {
	for {}
}

func (s *shadow) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, s.HomeParams(r))
}

func NewThing(id, model, name string) *merle.Thing {
	s := &shadow{}

	s.Init = s.init
	s.Run = s.run
	s.Home = s.home

	return s.InitThing(id, model, name)
}
