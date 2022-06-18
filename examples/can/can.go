// file: examples/can/can.go

package can

import (
	"github.com/go-daq/canbus"
	"github.com/merliot/merle"
	"log"
)

type can struct {
	Iface string
}

func NewCan() *can {
	return &can{
		Iface: "can0",
	}
}

type canMsg struct {
	Msg string
	Id  uint32
	Data []byte
}

func (c *can) run(p *merle.Packet) {
	sock, err := canbus.New()
	if err != nil {
		log.Println("Creating CAN bus failed:", err)
		return
	}

	err = sock.Bind(c.Iface)
	if err != nil {
		log.Printf("Binding to %s failed: %s", c.Iface, err)
		return
	}

	msg := &canMsg{Msg: "CAN"}

	for {
		msg.Id, msg.Data, err = sock.Recv()
		if err != nil {
			log.Println("Error reading CAN socket:", err)
			return
		}
		p.Marshal(&msg).Broadcast()
	}
}

func (c *can) getState(p *merle.Packet) {
	msg := merle.Msg{Msg: merle.ReplyState}
	p.Marshal(&msg).Reply()
}

func (c *can) can(p *merle.Packet) {
	if p.IsThing() {
		log.Println("Save CAN msg")
	}
	p.Broadcast()
}

func (c *can) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun:     c.run,
		merle.GetState:   c.getState,
		merle.ReplyState: nil,
		"CAN":            c.can,
	}
}
