package thermo

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/bmp180"
	"github.com/merliot/merle/examples/relays"
	"sync"
)

type thermo struct {
	sync.RWMutex
	relays struct {
		id     string
		online bool
	}
	bmp struct {
		id     string
		online bool
		temp   int
	}
}

func NewThermo() merle.Thinger {
	return &thermo{}
}

func (t *thermo) BridgeThingers() merle.BridgeThingers {
	return merle.BridgeThingers{
		".*:relays:.*": func() merle.Thinger { return relays.NewRelays() },
		".*:bmp180:.*": func() merle.Thinger { return bmp180.NewBmp180() },
	}
}

func (t *thermo) identity(p *merle.Packet) {
	var getState bool = false
	var msg merle.MsgIdentity
	p.Unmarshal(&msg)

	t.Lock()
	switch msg.Model {
	case "relays":
		if t.relays.id == "" || t.relays.id == msg.Id {
			t.relays.id = msg.Id
			t.relays.online = msg.Online
		}
	case "bmp180":
		if t.bmp.id == "" || t.bmp.id == msg.Id {
			t.bmp.id = msg.Id
			t.bmp.online = msg.Online
			if msg.Online {
				getState = true
			}
		}
	}
	t.Unlock()

	if (getState) {
		merle.ReplyGetState(p)
	}
}

func (t *thermo) update(p *merle.Packet) {
	var msg bmp180.MsgState
	p.Unmarshal(&msg)

	var on bool = (msg.Temperature > 76 && msg.Temperature <= 80)

	t.RLock()
	defer t.RUnlock()

	if p.Src() != t.bmp.id {
		return
	}

	t.bmp.temp = msg.Temperature

	if t.relays.online {
		msg := relays.MsgClick{
			Msg: "Click",
			Relay: 0,
			State: on,
		}
		p.Marshal(&msg)
		p.Send(t.relays.id)
	}
}

func (t *thermo) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.EventStatus:   merle.ReplyGetIdentity,
		merle.ReplyIdentity: t.identity,
		merle.ReplyState:    t.update,
		"Update":            t.update,
		"default":           nil, // drop everything else silently
	}
}

func (t *thermo) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:     merle.NoInit,
		merle.CmdRun:      merle.RunForever,
		merle.EventStatus: nil,
	}
}

func (t *thermo) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir:    "examples/thermo/assets",
		HtmlTemplate: "templates/thermo.html",
	}
}
