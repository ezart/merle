package can

import (
	"github.com/merliot/merle"
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

func (b *bridge) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		"CAN": merle.Broadcast, // broadcast CAN msgs to everyone
		"default": nil,         // drop everything else
	}
}

func (b *bridge) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: merle.RunForever,
	}
}

func (b *bridge) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{}
}
