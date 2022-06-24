// file: examples/tutorial/blinkv2/blink.go

package main

import (
	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
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
		merle.CmdInit: nil,
		merle.CmdRun: b.run,
	}
}

func (b *blink) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{}
}

func main() {
	thing := merle.NewThing(&blink{})
	log.Fatalln(thing.Run())
}
