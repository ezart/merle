package blink

import (
	"log"
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

type blink struct {
	demo      bool
	adaptor   *raspi.Adaptor
	led       *gpio.LedDriver
	lastState bool
	paused    bool
}

func NewBlinker(demo bool) merle.Thinger {
	return &blink{demo: demo}
}

type msgReplyPaused struct {
	Msg    string
	Paused bool
	State  bool
}

func (b *blink) sendPaused(p *merle.Packet) {
	msg := msgReplyPaused{
		Msg:    "ReplyPaused",
		Paused: b.paused,
		State:  b.lastState,
	}
	p.Marshal(&msg).Reply()
}

func (b *blink) savePaused(p *merle.Packet) {
	var msg msgReplyPaused
	p.Unmarshal(&msg)
	b.paused = msg.Paused
	b.lastState = msg.State
}

func (b *blink) pause(p *merle.Packet) {
	b.paused = true
	p.Broadcast()
}

func (b *blink) resume(p *merle.Packet) {
	b.paused = false
	p.Broadcast()
}

func (b *blink) runPrime(p *merle.Packet) {
	log.Println("****************** runPrime")
	msg := struct{ Msg string }{Msg: "GetPaused"}
	p.Marshal(&msg).Reply()
}

type spamLedState struct {
	Msg   string
	State bool
}

func (b *blink) ledState(p *merle.Packet) {
	var spam spamLedState
	p.Unmarshal(&spam)
	b.lastState = spam.State
	p.Broadcast()
}

func (b *blink) state() bool {
	if b.demo {
		return b.lastState
	}
	return b.led.State()
}

func (b *blink) toggle() {
	b.lastState = !b.lastState
	if !b.demo {
		b.led.Toggle()
	}
}

func (b *blink) sendLedState(p *merle.Packet) {
	spam := spamLedState{
		Msg:   "SpamLedState",
		State: b.state(),
	}
	p.Marshal(&spam).Broadcast()
}

func (b *blink) run(p *merle.Packet) {
	b.adaptor = raspi.NewAdaptor()
	b.adaptor.Connect()

	b.led = gpio.NewLedDriver(b.adaptor, "11")
	b.led.Start()
	b.lastState = b.led.State()

	ticker := time.NewTicker(time.Second)

	b.sendLedState(p)

	for {
		select {
		case <-ticker.C:
			if !b.paused {
				b.toggle()
				b.sendLedState(p)
			}
		}
	}
}

func (b *blink) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"_CmdRunPrime", b.runPrime},
		{"_CmdRun", b.run},
		{"GetPaused", b.sendPaused},
		{"ReplyPaused", b.savePaused},
		{"CmdPause", b.pause},
		{"CmdResume", b.resume},
		{"SpamLedState", b.ledState},
	}
}

func (b *blink) Config(config merle.Configurator) error {
	return nil
}

func (b *blink) Template() string {
	return "examples/blink/assets/templates/blink.html"
}
