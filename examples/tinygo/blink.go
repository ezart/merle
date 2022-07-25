// file: examples/tinygo/blink.go

//go:build tinygo
// +build tinygo

// tinygo flash -target=arduino-nano33 examples/tinygo/blink.go

package main

import (
	"machine"
	"time"

	"github.com/merliot/merle"
//	"tinygo.org/x/drivers/net"
)

// Access point info
const ssid = ""
const pass = ""

type blinky struct {
}

func (b *blinky) init(p *merle.Packet) {
	merle.Nano33ConnectAP(ssid, pass)
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
		merle.CmdInit: b.init,
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
