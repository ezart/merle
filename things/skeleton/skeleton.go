// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Skeleton is a bare bones example of a thing
package skeleton

import (
	"github.com/scottfeldman/merle"
	"time"
	"net/http"
)

type skeleton struct {
	merle.Thing
}

func (s *skeleton) init() error {
	return nil
}

func (s *skeleton) run() {
	for {}
}

func (s *skeleton) home(w http.ResponseWriter, r *http.Request) {
}

func (s *skeleton) animate(p *merle.Packet) {
}

func NewThing(id, model, name string) *merle.Thing {
	s := skeleton{}

	s.Status = "online"
	s.Id = id
	s.Model = model
	s.Name = name
	s.StartupTime = time.Now()

	s.Init = s.init
	s.Run = s.run
	s.Home = s.home

	s.AddHandler("animate", s.animate)

	return &s.Thing
}
