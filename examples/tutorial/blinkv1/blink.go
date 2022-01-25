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

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			led.Toggle()
		}
	}
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: b.run,
	}
}

func (b *blink) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{}
}

func main() {
	var cfg merle.ThingConfig

	merle.NewThing(&blink{}, &cfg).Run()
}
