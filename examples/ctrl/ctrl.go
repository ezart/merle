package main

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/bmp180"
	"github.com/merliot/merle/examples/relays"
	"log"
	"sync"
)

type ctrl struct {
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

func (c *ctrl) BridgeThingers() merle.BridgeThingers {
	return merle.BridgeThingers{
		".*:relays:.*": func() merle.Thinger { return relays.NewRelays() },
		".*:bmp180:.*": func() merle.Thinger { return bmp180.NewBmp180() },
	}
}

func (c *ctrl) identity(p *merle.Packet) {
	var getState bool = false
	var msg merle.MsgIdentity
	p.Unmarshal(&msg)

	c.Lock()
	switch msg.Model {
	case "relays":
		if c.relays.id == "" || c.relays.id == msg.Id {
			c.relays.id = msg.Id
			c.relays.online = msg.Online
		}
	case "bmp180":
		if c.bmp.id == "" || c.bmp.id == msg.Id {
			c.bmp.id = msg.Id
			c.bmp.online = msg.Online
			if msg.Online {
				getState = true
			}
		}
	}
	c.Unlock()

	if (getState) {
		merle.ReplyGetState(p)
	}
}

func (c *ctrl) update(p *merle.Packet) {
	var msg bmp180.MsgState
	p.Unmarshal(&msg)

	var on bool = (msg.Temperature > 76 && msg.Temperature <= 80)

	c.RLock()
	defer c.RUnlock()

	if p.Src() != c.bmp.id {
		return
	}

	c.bmp.temp = msg.Temperature

	if c.relays.online {
		msg := relays.MsgClick{
			Msg: "Click",
			Relay: 0,
			State: on,
		}
		p.Marshal(&msg)
		p.Send(c.relays.id)
	}
}

func (c *ctrl) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.EventStatus:   merle.ReplyGetIdentity,
		merle.ReplyIdentity: c.identity,
		merle.ReplyState:    c.update,
		"Update":            c.update,
		"default":           nil, // drop everything else silently
	}
}

func (c *ctrl) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:     merle.NoInit,
		merle.CmdRun:      merle.RunForever,
		merle.EventStatus: nil,
	}
}

const html = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body style="background-color:orange">
		
	</body>
</html>`

func (c *ctrl) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		HtmlTemplateText: html,
	}
}

func main() {
	thing := merle.NewThing(&ctrl{})

	thing.Cfg.Model = "ctrl"
	thing.Cfg.Name = "controller"
	thing.Cfg.User = "merle"

	thing.Cfg.PortPublic = 80
	thing.Cfg.PortPublicTLS = 443
	thing.Cfg.PortPrivate = 8080

	log.Fatalln(thing.Run())
}
