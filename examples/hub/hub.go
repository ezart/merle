package hub

import (
	"github.com/merliot/merle"
	"github.com/merliot/merle/examples/relays"
)

type hub struct {
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

func (h *hub) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun: merle.RunForever,
	}
}
