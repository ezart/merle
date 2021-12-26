// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package raspi_blink

import (
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"html/template"
	"net/http"
	"time"
)

var templ *template.Template

func init() {
	templ = template.Must(template.ParseFiles("web/templates/raspi_blink.html"))
}

type blinker struct {
	merle.Thing
	adaptor *raspi.Adaptor
	led     *gpio.LedDriver
	paused  bool
}

func (b *blinker) init() error {
	b.adaptor = raspi.NewAdaptor()
	b.adaptor.Connect()

	b.led = gpio.NewLedDriver(b.adaptor, "11")
	b.led.Start()

	return nil
}

func (b *blinker) sendState() {
	msg := struct {
		Msg   string
		State bool
	}{
		Msg:   "state",
		State: b.led.State(),
	}
	b.Broadcast(merle.NewPacket(&msg))
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

func (b *blinker) getPaused(p *merle.Packet) {
	msg := struct {
		Msg    string
		Paused bool
	}{
		Msg:    "paused",
		Paused: b.paused,
	}
	b.Reply(merle.UpdatePacket(p, &msg))

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
	b := &blinker{}

	b.Init = b.init
	b.Run = b.run
	b.Home = b.home

	b.HandleMsg("paused", b.getPaused)
	b.HandleMsg("pause", b.pause)
	b.HandleMsg("resume", b.resume)
	b.HandleMsg("state", b.Broadcast)

	return b.InitThing(id, model, name)
}
