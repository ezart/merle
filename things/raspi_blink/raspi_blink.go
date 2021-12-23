// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package raspi_blink

import (
	"html/template"
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
	"net/http"
)

var templ *template.Template

func init() {
	templ = template.Must(template.ParseFiles("web/templates/raspi_blink.html"))
}

type blinker struct {
	merle.Thing
	adaptor *raspi.Adaptor
	led *gpio.LedDriver
	paused bool
}

func (b *blinker) init() error {
	b.adaptor = raspi.NewAdaptor()
	b.adaptor.Connect()

	b.led = gpio.NewLedDriver(b.adaptor, "7")
	b.led.Start()

	return nil
}

type msgState struct {
	Type string
	State bool
}

func (b *blinker) sendState() {
	var msg = msgState{
		Type:  "state",
		State: b.led.State(),
	}
	b.Broadcast(b.NewPacket(&msg))
}

func (b *blinker) run() {
	ticker := time.NewTicker(time.Second)

	b.sendState()

	for {
		select {
		case <-ticker.C:
			if !b.paused {
				b.led.Toggle()
				b.sendState()
			}
		}
	}
}

func (b *blinker) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, b.HomeParams(r))
}

func (b *blinker) pause(p *merle.Packet) {
	b.paused = true
	b.Sink(p)
	b.Broadcast(p)
}

func (b *blinker) resume(p *merle.Packet) {
	b.paused = false
	b.Sink(p)
	b.Broadcast(p)
}

func NewThing(id, model, name string) *merle.Thing {
	b := blinker{}

	if id == "" {
		id = merle.DefaultId_()
	}

	b.Status = "online"
	b.Id = id
	b.Model = model
	b.Name = name
	b.StartupTime = time.Now()

	b.Init = b.init
	b.Run = b.run
	b.Home = b.home

	b.AddHandler("pause", b.pause)
	b.AddHandler("resume", b.resume)
	b.AddHandler("state", b.Broadcast)

	return &b.Thing
}
