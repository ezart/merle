// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package factory

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/things/raspi_blink"
	"github.com/scottfeldman/merle/things/skeleton"
	"github.com/scottfeldman/merle/things/hub"
)

var things = map[string]func(id, model, name string) *merle.Thing{
	"raspi_blink": raspi_blink.NewThing,
	"skeleton": skeleton.NewThing,
	"hub": hub.NewThing,
}

func NewThing(id, model, name string) *merle.Thing {

	if f, ok := things[model]; ok {
		return f(id, model, name)
	}

	return nil
}
