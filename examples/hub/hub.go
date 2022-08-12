package hub

import (
	"sync"

	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/bmp180"
	"github.com/merliot/merle/examples/gps"
	"github.com/merliot/merle/examples/relays"
)

type child struct {
	Id     string
	Online bool
}

type hub struct {
	sync.Mutex
	Msg      string
	Children map[string]child
}

func NewHub() merle.Thinger {
	return &hub{Msg: merle.ReplyState}
}

func (h *hub) BridgeThingers() merle.BridgeThingers {
	return merle.BridgeThingers{
		".*:relays:.*": func() merle.Thinger { return relays.NewRelays() },
		".*:gps:.*":    func() merle.Thinger { return gps.NewGps() },
		".*:bmp180:.*": func() merle.Thinger { return bmp180.NewBmp180() },
	}
}

func (h *hub) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		"default": nil, // drop everything silently
	}
}

func (h *hub) update(p *merle.Packet) {
	var msg merle.MsgEventStatus
	p.Unmarshal(&msg)

	child := child{
		Id:     msg.Id,
		Online: msg.Online,
	}

	h.Lock()
	h.Children[msg.Id] = child
	h.Unlock()

	p.Broadcast()
}

func (h *hub) getState(p *merle.Packet) {
	h.Lock()
	p.Marshal(h)
	h.Unlock()
	p.Reply()
}

func (h *hub) init(p *merle.Packet) {
	h.Children = make(map[string]child)
}

func (h *hub) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit:     h.init,
		merle.CmdRun:      merle.RunForever,
		merle.GetState:    h.getState,
		merle.EventStatus: h.update,
	}
}

func (h *hub) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir:    "examples/hub/assets",
		HtmlTemplate: "templates/hub.html",
	}
}
