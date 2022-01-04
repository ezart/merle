// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Skeleton is a bare bones example of a Thing
package skeleton

import (
	"github.com/scottfeldman/merle"
	"net/http"
)

type skeleton struct {
	merle.Thing
}

func (s *skeleton) animate(p *merle.Packet) {
}

func (s *skeleton) init() error {
	return nil
}

func (s *skeleton) run() {
	for {
	}
}

func (s *skeleton) home(w http.ResponseWriter, r *http.Request) {
}

func NewSkeleton(id, model, name string) *merle.Thing {
	s := &skeleton{}

	t := s.InitThing(id, model, name)
	if t == nil {
		return nil
	}

	t.Init = s.init
	t.Run = s.run
	t.Home = s.home

	t.Subscribe("CmdAnimate", s.animate)

	return t
}
