// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package stork

import (
	"fmt"
	"github.com/scottfeldman/merle"
	"github.com/scottfeldman/merle/things/chat"
	"github.com/scottfeldman/merle/things/hub"
	"github.com/scottfeldman/merle/things/raspi_blink"
	"github.com/scottfeldman/merle/things/hello_world"
	"github.com/scottfeldman/merle/things/test"
	"github.com/scottfeldman/merle/things/prime"
	glog "log"
)

type stork struct {
}

func NewStork() merle.Storker {
	return &stork{}
}

func (s *stork) NewThinger(log *glog.Logger, model string, demo bool) (merle.Thinger, error) {

	var thingers = map[string]func(*glog.Logger, bool) merle.Thinger{
		"test":        test.NewModel,
		"hello_world": hello_world.NewModel,
		"raspi_blink": raspi_blink.NewModel,
		"hub":         hub.NewModel,
		"chat":        chat.NewModel,
		"prime":       prime.NewModel,
	}

	if thinger, ok := thingers[model]; ok {
		return thinger(log, demo), nil
	}

	return nil, fmt.Errorf("Model '%s' unknown", model)
}
