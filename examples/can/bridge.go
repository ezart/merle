// file: examples/can/bridge.go

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
		".*:can_node:.*": func() merle.Thinger { return NewNode() },
	}
}

func (b *bridge) BridgeSubscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.EventBridgeConnect: merle.ReplyGetState,
		"CAN": merle.Broadcast, // broadcast CAN msgs to everyone
		"default": nil,         // drop everything else silently
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
