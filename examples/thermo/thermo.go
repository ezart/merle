package thermo

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/bmp180"
	"github.com/merliot/merle/examples/relays"
	"sync"
	"log"
)

type thermo struct {
	sync.RWMutex
	refresh chan bool
	Msg string
	Relays struct {
		Id     string
		Online bool
		States [4]bool
	}
	Sensors struct {
		Id       string
		Online   bool
		Temp     int
		prevTemp int
	}
	SetPoint int
}

func NewThermo() merle.Thinger {
	return &thermo{Msg: merle.ReplyState}
}

func (t *thermo) BridgeThingers() merle.BridgeThingers {
	return merle.BridgeThingers{
		".*:relays:.*": func() merle.Thinger { return relays.NewRelays() },
		".*:bmp180:.*": func() merle.Thinger { return bmp180.NewBmp180() },
	}
}

func (t *thermo) relayClick(p *merle.Packet, relay int, on bool) {
	msg := relays.MsgClick{
		Msg: "Click",
		Relay: relay,
		State: on,
	}
	p.Marshal(&msg)
	log.Println("SENDING CLICK")
	p.Send(t.Relays.Id)
	log.Println("SENT CLICK")
}

func (t *thermo) calculate(p *merle.Packet) {
	var furnaceClick bool = false
	var airCondClick bool = false
	var tempChanged bool = false

	t.Lock()

	tempChanged = (t.Sensors.Temp != t.Sensors.prevTemp)
	t.Sensors.prevTemp = t.Sensors.Temp

	// Turn furnace relay on if current temp < setpoint,
	// other turn on air conditioner relay

	var wantFurnaceOn bool = (t.Sensors.Temp < t.SetPoint)
	var wantAirCondOn bool = !wantFurnaceOn

	var furnaceRelay *bool = &t.Relays.States[0]
	var airCondRelay *bool = &t.Relays.States[1]

	if *furnaceRelay != wantFurnaceOn {
		*furnaceRelay = wantFurnaceOn
		furnaceClick = t.Relays.Online
	}

	if *airCondRelay != wantAirCondOn {
		*airCondRelay = wantAirCondOn
		airCondClick = t.Relays.Online
	}

	t.Unlock()

	if furnaceClick {
		t.relayClick(p, 0, wantFurnaceOn)
	}
	if airCondClick {
		t.relayClick(p, 1, wantAirCondOn)
	}

	log.Println("REFRESH", furnaceClick, airCondClick, tempChanged)
	if (furnaceClick || airCondClick || tempChanged) {
		t.refresh <- true
	}
}

func (t *thermo) identity(p *merle.Packet) {
	var calculate bool = false
	var msg merle.MsgIdentity
	p.Unmarshal(&msg)

	t.Lock()
	switch msg.Model {
	case "relays":
		if t.Relays.Id == "" || t.Relays.Id == msg.Id {
			t.Relays.Id = msg.Id
			t.Relays.Online = msg.Online
			calculate = true
		}
	case "bmp180":
		if t.Sensors.Id == "" || t.Sensors.Id == msg.Id {
			t.Sensors.Id = msg.Id
			t.Sensors.Online = msg.Online
			calculate = true
		}
	}
	t.Unlock()

	if (calculate) {
		if msg.Online {
			merle.ReplyGetState(p)
		} else {
			t.calculate(p)
		}
	}
}

func (t *thermo) state(p *merle.Packet) {
	switch p.Src() {
	case t.Relays.Id:
		var msg relays.MsgState
		p.Unmarshal(&msg)
		t.Lock()
		t.Relays.States = msg.States
		t.Unlock()
	case t.Sensors.Id:
		var bmp bmp180.Bmp180
		p.Unmarshal(&bmp)
		t.Lock()
		t.Sensors.Temp = bmp.Temperature
		t.Unlock()
	}

	t.calculate(p)
}

func (t *thermo) update(p *merle.Packet) {
	switch p.Src() {
	case t.Sensors.Id:
		var bmp bmp180.Bmp180
		p.Unmarshal(&bmp)
		t.Lock()
		t.Sensors.Temp = bmp.Temperature
		t.Unlock()
	default:
		return
	}

	t.calculate(p)
}

func (t *thermo) click(p *merle.Packet) {
	switch p.Src() {
	case t.Relays.Id:
		var msg relays.MsgClick
		p.Unmarshal(&msg)
		t.Lock()
		t.Relays.States[msg.Relay] = msg.State
		t.Unlock()
	default:
		return
	}

	t.calculate(p)
}

func (t *thermo) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.EventStatus:   merle.ReplyGetIdentity,
		merle.ReplyIdentity: t.identity,
		merle.ReplyState:    t.state,
		"Update":            t.update, // from bmp180
		"Click":             t.click,  // from relays
		"default":           nil,      // drop everything else silently
	}
}

func (t *thermo) init(p *merle.Packet) {
	t.refresh = make(chan bool)
	t.SetPoint = 68 // Nixon
}

func (t *thermo) marshal(p *merle.Packet) {
	t.RLock()
	p.Marshal(t)
	t.RUnlock()
}

func (t *thermo) run(p *merle.Packet) {
	for {
		select {
		case <- t.refresh:
			log.Println("GOT REFRESH")
			t.marshal(p)
			p.Broadcast()
		}
	}
}

func (t *thermo) getState(p *merle.Packet) {
	log.Println("GETSTATE")
	t.marshal(p)
	p.Reply()
}

func (t *thermo) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:     t.init,
		merle.CmdRun:      t.run,
		merle.GetState:    t.getState,
		merle.EventStatus: nil,
	}
}

func (t *thermo) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir:    "examples/thermo/assets",
		HtmlTemplate: "templates/thermo.html",
	}
}
