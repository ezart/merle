// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package stork

import (
	"fmt"
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/things/test"
	"github.com/scottfeldman/merle/things/skeleton"
	"github.com/scottfeldman/merle/things/raspi_blink"
	//	"github.com/scottfeldman/merle/things/chat"
	//	"github.com/scottfeldman/merle/things/bridge"
	//	"github.com/scottfeldman/merle/things/hub"
)

var models = map[string]func(demo bool) merle.IThing{
	"test":        test.NewModel,
	"skeleton":    skeleton.NewModel,
	"raspi_blink": raspi_blink.NewModel,
	//	"chat":        chat.NewChat,
	//	"bridge":      bridge.NewBridge,
	//	"hub":         hub.NewHub,
}

func NewModel(model string, demo bool) (merle.IThing, error) {

	if f, ok := models[model]; ok {
		return f(demo), nil
	}

	return nil, fmt.Errorf("Model '%s' unknown", model)
}
