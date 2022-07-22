//go:build tinygo
// +build tinygo

package main

import (
	"github.com/merliot/merle"
	"machine"
	"time"
)

type blinky struct {
}

func (b *blinky) run(p *merle.Packet) {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		led.Low()
		time.Sleep(time.Millisecond * 500)

		led.High()
		time.Sleep(time.Millisecond * 500)
	}
}

func (b *blinky) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit: nil,
		merle.CmdRun:  b.run,
	}
}

func (b *blinky) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{}
}

func main() {
	thing := merle.NewThing(&blinky{})
	thing.Run()
}
