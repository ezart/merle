// file: examples/can/node.go

package can

import (
	"github.com/go-daq/canbus"
	"github.com/merliot/merle"
	"log"
)

type node struct {
	Iface string
	sock *canbus.Socket
}

func NewCan() *node {
	return &node{Iface: "can0"}
}

type canMsg struct {
	Msg string
	Id  uint32
	Data []byte
}

func (n *node) run(p *merle.Packet) {
	var err error

	n.sock, err = canbus.New()
	if err != nil {
		log.Println("Creating CAN bus failed:", err)
		return
	}

	err = n.sock.Bind(n.Iface)
	if err != nil {
		log.Printf("Binding to %s failed: %s", n.Iface, err)
		return
	}

	msg := &canMsg{Msg: "CAN"}

	for {
		msg.Id, msg.Data, err = n.sock.Recv()
		if err != nil {
			log.Println("Error reading CAN socket:", err)
			return
		}
		p.Marshal(&msg).Broadcast()
	}
}

func (n *node) can(p *merle.Packet) {
	if p.IsThing() {
		var msg canMsg

		p.Unmarshal(&msg)
		_, err := n.sock.Send(msg.Id, msg.Data)
		if err != nil {
			log.Println("Error writing CAN socket:", err)
		}
	}
	p.Broadcast()
}

func (n *node) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun:     n.run,
		merle.GetState:   merle.ReplyStateEmpty,
		merle.ReplyState: nil,
		"CAN":            n.can,
	}
}
