package raspi_blink

import (
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

type blinker struct {
	merle.Thing
	adaptor *raspi.Adaptor
	led *gpio.LedDriver
}

func (b *blinker) init() error {
	b.adaptor = raspi.NewAdaptor()
	b.adaptor.Connect()

	b.led = gpio.NewLedDriver(b.adaptor, "7")
	b.led.Start()

	return nil
}

func (b *blinker) run() {
	for {
		gobot.Every(1*time.Second, func() {
			b.led.Toggle()
		})
	}
}

func NewThing(name string) *merle.Thing {
	b := blinker{}

	b.Status = "online"
	b.Id = merle.DefaultId_()
	b.Model = "blinker"
	b.Name = name
	b.StartupTime = time.Now()

	b.Init = b.init
	b.Run = b.run

	return &b.Thing
}
