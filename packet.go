// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"encoding/json"
)

// A Packet contains a JSON message and a source connection.
type Packet struct {
	bus *bus
	src socketer
	msg []byte
}

func newPacket(bus *bus, src socketer, msg interface{}) *Packet {
	p := &Packet{bus: bus, src: src}
	p.msg, _ = json.Marshal(msg)
	return p
}

func (p *Packet) clone(bus *bus, src socketer) *Packet {
	return &Packet{bus: bus, src: src, msg: p.msg}
}

func (p *Packet) Marshal(msg interface{}) *Packet {
	p.msg, _ = json.Marshal(msg)
	return p
}

func (p *Packet) Unmarshal(msg interface{}) {
	json.Unmarshal(p.msg, msg)
}

func (p *Packet) String() string {
	return string(p.msg)
}

func (p *Packet) Source() interface{} {
	return p.src
}

func (p *Packet) Send(dst interface{}) {
	dstSock, ok := dst.(socketer)
	if ok {
		if err := p.bus.send(p, dstSock); err != nil {
			p.bus.log.Println(err)
		} else {
			p.bus.log.Printf("Send: %.80s", p.String())
		}
	} else {
		p.bus.log.Println("Send: can't send to non-socket")
	}
}

func (p *Packet) Reply() {
	if err := p.bus.reply(p); err != nil {
		p.bus.log.Println(err)
	} else {
		p.bus.log.Printf("Reply: %.80s", p.String())
	}
}

func (p *Packet) Broadcast() {
	if err := p.bus.broadcast(p); err != nil {
		p.bus.log.Println(err)
	} else {
		p.bus.log.Printf("Broadcast: %.80s", p.String())
	}
}
