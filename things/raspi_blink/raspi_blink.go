package raspi_blink

import (
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"time"
)

type raspi_blink struct {
	demo      bool
	adaptor   *raspi.Adaptor
	led       *gpio.LedDriver
	lastState bool
	paused    bool
}

func NewModel(l *log.Logger, demo bool) merle.Thinger {
	return &raspi_blink{demo: demo}
}

type msgReplyPaused struct {
	Msg    string
	Paused bool
	State  bool
}

func (r *raspi_blink) sendPaused(p *merle.Packet) {
	msg := msgReplyPaused{
		Msg:    "ReplyPaused",
		Paused: r.paused,
		State:  r.lastState,
	}
	p.Marshal(&msg).Reply()
}

func (r *raspi_blink) savePaused(p *merle.Packet) {
	var msg msgReplyPaused
	p.Unmarshal(&msg)
	r.paused = msg.Paused
	r.lastState = msg.State
}

func (r *raspi_blink) pause(p *merle.Packet) {
	r.paused = true
	p.Broadcast()
}

func (r *raspi_blink) resume(p *merle.Packet) {
	r.paused = false
	p.Broadcast()
}

func (r *raspi_blink) start(p *merle.Packet) {
	msg := struct{ Msg string }{Msg: "GetPaused"}
	p.Marshal(&msg).Reply()
}

type spamLedState struct {
	Msg   string
	State bool
}

func (r *raspi_blink) ledState(p *merle.Packet) {
	var spam spamLedState
	p.Unmarshal(&spam)
	r.lastState = spam.State
	p.Broadcast()
}

func (r *raspi_blink) state() bool {
	if r.demo {
		return r.lastState
	}
	return r.led.State()
}

func (r *raspi_blink) toggle() {
	r.lastState = !r.lastState
	if !r.demo {
		r.led.Toggle()
	}
}

func (r *raspi_blink) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"GetPaused", r.sendPaused},
		{"ReplyPaused", r.savePaused},
		{"CmdPause", r.pause},
		{"CmdResume", r.resume},
		{"CmdStart", r.start},
		{"SpamLedState", r.ledState},
	}
}

func (r *raspi_blink) Config(config merle.Configurator) error {
	return nil
}

func (r *raspi_blink) Template() string {
	return "web/templates/raspi_blink.html"
}

func (r *raspi_blink) sendLedState(p *merle.Packet) {
	spam := spamLedState{
		Msg:   "SpamLedState",
		State: r.state(),
	}
	p.Marshal(&spam).Broadcast()
}

func (r *raspi_blink) Run(p *merle.Packet) {
	r.adaptor = raspi.NewAdaptor()
	r.adaptor.Connect()

	r.led = gpio.NewLedDriver(r.adaptor, "11")
	r.led.Start()
	r.lastState = r.led.State()

	ticker := time.NewTicker(time.Second)

	r.sendLedState(p)

	for {
		select {
		case <-ticker.C:
			if !r.paused {
				r.toggle()
				r.sendLedState(p)
			}
		}
	}
}
