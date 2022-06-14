// file: examples/can/can.go

package main

import (
	"flag"
	"github.com/go-daq/canbus"
	"github.com/merliot/merle"
	"log"
)

type can struct {
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

	iface := "can0"

	err = sock.Bind(iface)
	if err != nil {
		log.Printf("Binding to %s failed: %s", iface, err)
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
	msg := struct{ Msg string }{Msg: merle.ReplyState}
	p.Marshal(&msg).Reply()
}

func (c *can) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdRun:     c.run,
		merle.GetState:   c.getState,
		merle.ReplyState: nil,
	}
}

func (c *can) Assets() *merle.ThingAssets {
	return &merle.ThingAssets{}
}

func main() {
	thing := merle.NewThing(&can{})

	thing.Cfg.Model = "can"
	thing.Cfg.Name = "can0"
	thing.Cfg.User = "merle"

	flag.StringVar(&thing.Cfg.MotherHost, "rhost", "", "Remote host")
	flag.StringVar(&thing.Cfg.MotherUser, "ruser", "merle", "Remote user")
	flag.BoolVar(&thing.Cfg.IsPrime, "prime", false, "Run as Thing Prime")
	flag.UintVar(&thing.Cfg.PortPublicTLS, "TLS", 0, "TLS port")

	flag.Parse()

	log.Fatalln(thing.Run())
}
