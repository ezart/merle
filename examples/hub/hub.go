package hub

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/relays"
	"sync"
)

type child struct {
	Id string
	Model string
	Name string
	Connected bool
}

type hub struct {
	sync.RWMutex
	children map[string]child
	update chan child
}

func NewHub() merle.Thinger {
	return &hub{}
}

func (h *hub) BridgeThingers() merle.BridgeThingers {
	return merle.BridgeThingers{
		".*:relays:.*": func() merle.Thinger { return relays.NewThing() },
	}
}

func (h *hub) connect(p *merle.Packet) {
	var msg merle.MsgIdentity
	p.Unmarshal(&msg)

	child := child{
		Id: msg.Id,
		Model: msg.Model,
		Name: msg.Name,
		Connected: true,
	}

	h.Lock()
	h.children[msg.Id] = child
	h.Unlock()

	h.update <- child
}

func (h *hub) disconnect(p *merle.Packet) {
	h.Lock()
	child := h.children[p.Src()]
	child.Connected = false
	h.Unlock()

	h.update <- child
}

func (h *hub) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdBridgeConnect: merle.ReplyGetIdentity,
		merle.CmdBridgeDisconnect: h.disconnect,
		merle.ReplyIdentity: h.connect,
		"default": nil, // drop everything else silently
	}
}

type msgChildren struct {
	Msg string
	Children map[string]child
}

type msgUpdate struct {
	Msg string
	Child child
}

func (h *hub) getState(p *merle.Packet) {
	h.RLock()
	msg := &msgChildren{Msg: merle.ReplyState, Children: h.children}
	h.RUnlock()

	p.Marshal(&msg).Reply()
}

func (h *hub) init(p *merle.Packet) {
	h.children = make(map[string]child)
	h.update = make(chan child)
}

func (h *hub) run(p *merle.Packet) {
	for {
		select {
		case child := <- h.update:
			msg := msgUpdate{Msg: "Update", Child: child}
			p.Marshal(&msg).Broadcast()
		}
	}
}

func (h *hub) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit: h.init,
		merle.CmdRun: h.run,
		merle.GetState: h.getState,
	}
}

func (h *hub) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir: "examples/hub/assets",
		HtmlTemplate: "templates/hub.html",
	}
}
