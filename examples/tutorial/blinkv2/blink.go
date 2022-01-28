// file: examples/tutorial/blinkv2/blink.go

package main

import (
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

type blink struct {
}

func (b *blink) run(p *merle.Packet) {
	adaptor := raspi.NewAdaptor()
	adaptor.Connect()

	led := gpio.NewLedDriver(adaptor, "11")
	led.Start()

	for {
		led.Toggle()
		time.Sleep(time.Second)
	}
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: b.run,
	}
}

func main() {
	var cfg merle.ThingConfig

	merle.NewThing(&blink{}, &cfg).Run()
}
