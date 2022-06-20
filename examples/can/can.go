// file: examples/can/can.go

package can

import (
	"github.com/go-daq/canbus"
	"github.com/merliot/merle"
	"log"
)

type can struct {
	Iface string
	sock *canbus.Socket
}

func NewCan() *can {
	return &can{Iface: "can0"}
}

type canMsg struct {
	Msg string
	Id  uint32
	Data []byte
}

func (c *can) run(p *merle.Packet) {
	var err error

	c.sock, err = canbus.New()
	if err != nil {
		log.Println("Creating CAN bus failed:", err)
		return
	}

	err = c.sock.Bind(c.Iface)
	if err != nil {
		log.Printf("Binding to %s failed: %s", c.Iface, err)
		return
	}

	msg := &canMsg{Msg: "CAN"}

	for {
		msg.Id, msg.Data, err = c.sock.Recv()
		if err != nil {
			log.Println("Error reading CAN socket:", err)
			return
		}
		p.Marshal(&msg).Broadcast()
	}
}

func (c *can) can(p *merle.Packet) {
	if p.IsThing() {
		var msg canMsg

		p.Unmarshal(&msg)
		_, err := c.sock.Send(msg.Id, msg.Data)
		if err != nil {
			log.Println("Error writing CAN socket:", err)
		}
	}
	p.Broadcast()
}

func (c *can) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun:     c.run,
		merle.GetState:   merle.ReplyStateEmpty,
		merle.ReplyState: nil,
		"CAN":            c.can,
	}
}
