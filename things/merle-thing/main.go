// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package main

import (
	"github.com/scottfeldman/merle/things/raspi_blink"
)

func main() {
	raspi_blink.NewThing("Blinky").Start()
}
