// file: examples/tutorial/blinkv1/blink.go

package main

import (
	"log"
	"time"

	"github.com/merliot/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

type blink struct {
	adaptor *raspi.Adaptor
	led     *gpio.LedDriver
}

func (b *blink) init(p *merle.Packet) {
	b.adaptor = raspi.NewAdaptor()
	b.adaptor.Connect()
	b.led = gpio.NewLedDriver(b.adaptor, "11")
	b.led.Start()
}

func (b *blink) run(p *merle.Packet) {
	for {
		b.led.Toggle()
		time.Sleep(time.Second)
	}
}

func (b *blink) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit: b.init,
		merle.CmdRun:  b.run,
	}
}

func (b *blink) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{}
}

func main() {
	thing := merle.NewThing(&blink{})
	log.Fatalln(thing.Run())
}
