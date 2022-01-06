// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package stork

import (
	"fmt"
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/things/skeleton"
	"github.com/scottfeldman/merle/things/raspi_blink"
	"github.com/scottfeldman/merle/things/chat"
	"github.com/scottfeldman/merle/things/bridge"
	"github.com/scottfeldman/merle/things/hub"
)

var things = map[string]func(id, model, name string) (*merle.Thing, error){
	"skeleton":    skeleton.NewSkeleton,
	"raspi_blink": raspi_blink.NewRaspiBlink,
	"chat":        chat.NewChat,
	"bridge":      bridge.NewBridge,
	"hub":         hub.NewHub,
}

func NewThing(id, model, name string) (*merle.Thing, error) {

	if f, ok := things[model]; ok {
		return f(id, model, name)
	}

	return nil, fmt.Errorf("Model '%s' unknown", model)
}
