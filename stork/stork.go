// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package stork

import (
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/things/bridge"
	"github.com/scottfeldman/merle/things/chat"
	"github.com/scottfeldman/merle/things/hub"
	"github.com/scottfeldman/merle/things/raspi_blink"
	"github.com/scottfeldman/merle/things/skeleton"
)

var things = map[string]func(id, model, name string) *merle.Thing{
	"skeleton":    skeleton.NewThing,
	"hub":         hub.NewThing,
	"bridge":      bridge.NewThing,
	"raspi_blink": raspi_blink.NewThing,
	"chat":        chat.NewThing,
}

func NewThing(id, model, name string) *merle.Thing {

	if f, ok := things[model]; ok {
		return f(id, model, name)
	}

	return nil
}
