// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
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
	lastState bool
	paused    bool
}

type msgReplyPaused struct {
	Msg    string
	Paused bool
	State  bool
}

func (b *blinker) sendPaused(p *merle.Packet) {
	msg := msgReplyPaused{
		Msg:    "ReplyPaused",
		Paused: b.paused,
		State: b.lastState,
	}
	p.Marshal(&msg).Reply()
}

func (b *blinker) savePaused(p *merle.Packet) {
	var msg msgReplyPaused
	p.Unmarshal(&msg)
	b.paused = msg.Paused
	b.lastState = msg.State
}

func (b *blinker) pause(p *merle.Packet) {
	b.paused = true
	p.Broadcast()
}

func (b *blinker) resume(p *merle.Packet) {
	b.paused = false
	p.Broadcast()
}

func (b *blinker) start(p *merle.Packet) {
	msg := struct{ Msg string }{Msg: "GetPaused"}
	p.Marshal(&msg).Reply()
}

type spamLedState struct {
	Msg   string
	State bool
}

func (b *blinker) ledState(p *merle.Packet) {
	var spam spamLedState
	p.Unmarshal(&spam)
	b.lastState = spam.State
	b.Broadcast(p)
}

func (b *blinker) init() error {
	b.adaptor = raspi.NewAdaptor()
	b.adaptor.Connect()

	b.led = gpio.NewLedDriver(b.adaptor, "11")
	b.led.Start()
	b.lastState = b.led.State()

	return nil
}

func (b *blinker) state() bool {
	if b.DemoMode() {
		return b.lastState
	}
	return b.led.State()
}

func (b *blinker) toggle() {
	b.lastState = !b.lastState
	if !b.DemoMode() {
		b.led.Toggle()
	}
}

func (b *blinker) sendLedState() {
	spam := spamLedState{
		Msg:   "SpamLedState",
		State: b.state(),
	}
	b.NewPacket(&spam).Broadcast()
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

func NewRaspiBlink(id, model, name string) *merle.Thing {
	b := &blinker{}

	t := b.InitThing(id, model, name)
	if t == nil {
		return nil
	}

	t.Init = b.init
	t.Run = b.run
	t.Home = b.home

	t.Subscribe("GetPaused", b.sendPaused)
	t.Subscribe("ReplyPaused", b.savePaused)
	t.Subscribe("CmdPause", b.pause)
	t.Subscribe("CmdResume", b.resume)
	t.Subscribe("CmdStart", b.start)
	t.Subscribe("SpamLedState", b.ledState)

	return t
}
