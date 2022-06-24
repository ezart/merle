package hub

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/relays"
)

type child struct {
	Id string
	Model string
	Name string
	Connected bool
}

type hub struct {
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

func (h *hub) connect(p *merle.Packet) {
	var msg merle.MsgIdentity
	p.Unmarshal(&msg)
	h.children[msg.Id] = child{
		Id: msg.Id,
		Model: msg.Model,
		Name: msg.Name,
		Connected: true,
	}
}

func (h *hub) disconnect(p *merle.Packet) {
	child := h.children[p.Src()]
	child.Connected = false
}

func (h *hub) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdBridgeConnect: merle.ReplyGetIdentity,
		merle.CmdBridgeDisconnect: h.disconnect,
		merle.ReplyIdentity: h.connect,
		"default": nil, // drop everything silently
	}
}

type msgChildren struct {
	Msg string
	Children map[string]child
}

func (h *hub) getState(p *merle.Packet) {
	msg := &msgChildren{Msg: merle.ReplyState, Children: h.children}
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
	}
}

func (h *hub) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{
		AssetsDir: "examples/hub/assets",
		HtmlTemplate: "templates/hub.html",
	}
}
