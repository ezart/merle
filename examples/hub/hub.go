package hub

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/relays"
	"sync"
)

type child struct {
	Id string
	Connected bool
}

type hub struct {
	sync.RWMutex
	children map[string]child
}

func NewHub() merle.Thinger {
	return &hub{}
}

func (h *hub) BridgeThingers() merle.BridgeThingers {
	return merle.BridgeThingers{
		".*:relays:.*": func() merle.Thinger { return relays.NewThing() },
	}
}

func (h *hub) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		"default": nil, // drop everything silently
	}
}

func (h *hub) update(p *merle.Packet, connected bool) {
	var msg merle.MsgId
	p.Unmarshal(&msg)

	child := child{
		Id: msg.Id,
		Connected: connected,
	}

	h.Lock()
	h.children[msg.Id] = child
	h.Unlock()

	p.Broadcast()
}

func (h *hub) connect(p *merle.Packet) {
	h.update(p, true)
}

func (h *hub) disconnect(p *merle.Packet) {
	h.update(p, false)
}

type msgState struct {
	Msg string
	Children map[string]child
}

func (h *hub) getState(p *merle.Packet) {
	h.RLock()
	msg := &msgState{Msg: merle.ReplyState, Children: h.children}
	h.RUnlock()

	p.Marshal(&msg).Reply()
}

func (h *hub) init(p *merle.Packet) {
	h.children = make(map[string]child)
}

func (h *hub) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit: h.init,
		merle.CmdRun: merle.RunForever,
		merle.GetState: h.getState,
		merle.EventConnect: h.connect,
		merle.EventDisconnect: h.disconnect,
	}
}

func (h *hub) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir: "examples/hub/assets",
		HtmlTemplate: "templates/hub.html",
	}
}
