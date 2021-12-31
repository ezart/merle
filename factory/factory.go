// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package factory

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/things/skeleton"
	"github.com/scottfeldman/merle/things/shadow"
	"github.com/scottfeldman/merle/things/hub"
	"github.com/scottfeldman/merle/things/raspi_blink"
	"github.com/scottfeldman/merle/things/chat"
)

var things = map[string]func(id, model, name string) *merle.Thing{
	"skeleton":    skeleton.NewThing,
	"hub":         hub.NewThing,
	"raspi_blink": raspi_blink.NewThing,
	"chat":        chat.NewThing,
}

func NewThing(id, model, name string) *merle.Thing {

	if f, ok := things[model]; ok {
		return f(id, model, name)
	}

	return nil
}
