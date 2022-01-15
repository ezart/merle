package raspi_blink

import (
	"github.com/scottfeldman/merle"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"time"
)

type thing struct {
	demo      bool
	adaptor   *raspi.Adaptor
	led       *gpio.LedDriver
	lastState bool
	paused    bool
}

func NewModel(log *log.Logger, demo bool) merle.Thinger {
	return &thing{demo: demo}
}

type msgReplyPaused struct {
	Msg    string
	Paused bool
	State  bool
}

func (t *thing) sendPaused(p *merle.Packet) {
	msg := msgReplyPaused{
		Msg:    "ReplyPaused",
		Paused: t.paused,
		State:  t.lastState,
	}
	p.Marshal(&msg).Reply()
}

func (t *thing) savePaused(p *merle.Packet) {
	var msg msgReplyPaused
	p.Unmarshal(&msg)
	t.paused = msg.Paused
	t.lastState = msg.State
}

func (t *thing) pause(p *merle.Packet) {
	t.paused = true
	p.Broadcast()
}

func (t *thing) resume(p *merle.Packet) {
	t.paused = false
	p.Broadcast()
}

func (t *thing) start(p *merle.Packet) {
	msg := struct{ Msg string }{Msg: "GetPaused"}
	p.Marshal(&msg).Reply()
}

type spamLedState struct {
	Msg   string
	State bool
}

func (t *thing) ledState(p *merle.Packet) {
	var spam spamLedState
	p.Unmarshal(&spam)
	t.lastState = spam.State
	p.Broadcast()
}

func (t *thing) state() bool {
	if t.demo {
		return t.lastState
	}
	return t.led.State()
}

func (t *thing) toggle() {
	t.lastState = !t.lastState
	if !t.demo {
		t.led.Toggle()
	}
}

func (t *thing) sendLedState(p *merle.Packet) {
	spam := spamLedState{
		Msg:   "SpamLedState",
		State: t.state(),
	}
	p.Marshal(&spam).Broadcast()
}

func (t *thing) run(p *merle.Packet) {
	t.adaptor = raspi.NewAdaptor()
	t.adaptor.Connect()

	t.led = gpio.NewLedDriver(t.adaptor, "11")
	t.led.Start()
	t.lastState = t.led.State()

	ticker := time.NewTicker(time.Second)

	t.sendLedState(p)

	for {
		select {
		case <-ticker.C:
			if !t.paused {
				t.toggle()
				t.sendLedState(p)
			}
		}
	}
}

func (t *thing) Subscribe() merle.Subscribers {
	return merle.Subscribers{
		{"CmdRun", t.run},
		{"GetPaused", t.sendPaused},
		{"ReplyPaused", t.savePaused},
		{"CmdPause", t.pause},
		{"CmdResume", t.resume},
		{"CmdStart", t.start},
		{"SpamLedState", t.ledState},
	}
}

func (t *thing) Config(config merle.Configurator) error {
	return nil
}

func (t *thing) Template() string {
	return "web/templates/raspi_blink.html"
}
