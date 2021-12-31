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
	adaptor   *raspi.Adaptor
	led       *gpio.LedDriver
	demoState bool
	paused    bool
}

func (b *blinker) init() error {
	b.adaptor = raspi.NewAdaptor()
	b.adaptor.Connect()

	b.led = gpio.NewLedDriver(b.adaptor, "11")
	b.led.Start()

	return nil
}

func (b *blinker) state() bool {
	if b.DemoMode() {
		return b.demoState
	}
	return b.led.State()
}

func (b *blinker) toggle() {
	if b.DemoMode() {
		b.demoState = !b.demoState
		return
	}
	b.led.Toggle()
}

func (b *blinker) sendLedState() {
	msg := struct {
		Msg   string
		State bool
	}{
		Msg:   "SpamLedState",
		State: b.state(),
	}
	b.Broadcast(merle.NewPacket(&msg))
}

func (b *blinker) run() {
	ticker := time.NewTicker(time.Second)

	b.sendLedState()

	for {
		select {
		case <-ticker.C:
			if !b.paused {
				b.toggle()
				b.sendLedState()
			}
		}
	}
}

func (b *blinker) home(w http.ResponseWriter, r *http.Request) {
	templ.Execute(w, b.HomeParams(r, nil))
}

type msgReplyPaused struct {
	Msg    string
	Paused bool
}

func (b *blinker) sendPaused(p *merle.Packet) {
	msg := msgReplyPaused{
		Msg:    "ReplyPaused",
		Paused: b.paused,
	}
	b.Reply(p.Marshal(&msg))

}

func (b *blinker) savePaused(p *merle.Packet) {
	var msg msgReplyPaused
	p.Unmarshal(&msg)
	b.paused = msg.Paused
}

func (b *blinker) pause(p *merle.Packet) {
	b.paused = true
	b.Broadcast(p)
}

func (b *blinker) resume(p *merle.Packet) {
	b.paused = false
	b.Broadcast(p)
}

func (b *blinker) start(p *merle.Packet) {
	msg := struct { Msg string }{ Msg: "GetPaused" }
	b.Reply(p.Marshal(&msg))
}

func NewThing(id, model, name string) *merle.Thing {
	b := &blinker{}

	b.Init = b.init
	b.Run = b.run
	b.Home = b.home

	b.Subscribe("GetPaused", b.sendPaused)
	b.Subscribe("ReplyPaused", b.savePaused)
	b.Subscribe("CmdPause", b.pause)
	b.Subscribe("CmdResume", b.resume)
	b.Subscribe("CmdStart", b.start)
	b.Subscribe("SpamLedState", b.Broadcast)

	return b.InitThing(id, model, name)
}
