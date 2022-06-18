package can

import (
	"github.com/merliot/merle"
	"log"
)

type bridge struct {
}

func NewBridge() merle.Thinger {
	return &bridge{}
}

func (b *bridge) BridgeThingers() merle.BridgeThingers {
	return merle.BridgeThingers{
		".*:can:.*": func() merle.Thinger { return NewCan() },
	}
}

func (b *bridge) connect(p *merle.Packet) {
	log.Println("bridge connected:", p.Id())
}

func (b *bridge) disconnect(p *merle.Packet) {
	log.Println("bridge disconnected:", p.Id())
}

func (b *bridge) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdBridgeConnect: b.connect,
		merle.CmdBridgeDisconnect: b.disconnect,
		"CAN": merle.Broadcast, // broadcast CAN msgs to everyone
		"default": nil,         // drop everything else
	}
}

func (b *bridge) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: merle.RunForever,
	}
}
