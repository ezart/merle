package main

import (
	"github.com/scottfeldman/merle/examples/raspi_blink"
)

func main() {
	t := raspi_blink.NewThing("Blinky")
	t.Start()
}
