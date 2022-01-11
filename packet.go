// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"encoding/json"
)

// A packet contains a message and a source.
type Packet struct {
	// Bus the packet lives on
	bus *bus
	// Source socket on bus; the packet source
	src socketer
	// Message
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

// JSON marshal into packet message
func (p *Packet) Marshal(msg interface{}) *Packet {
	p.msg, _ = json.Marshal(msg)
	return p
}

// JSON unmarshal from packet message
func (p *Packet) Unmarshal(msg interface{}) {
	json.Unmarshal(p.msg, msg)
}

// String representation of packet message
func (p *Packet) String() string {
	return string(p.msg)
}

func (p *Packet) Source() interface{} {
	return p.src
}

func (p *Packet) Send(dst interface{}) {
	p.bus.log.Printf("Send: %.80s", p.String())
	dstSock, ok := dst.(socketer)
	if ok {
		if err := p.bus.send(p, dstSock); err != nil {
			p.bus.log.Println(err)
		}
	} else {
		p.bus.log.Println("Send: can't send to non-socket")
	}
}

func (p *Packet) Reply() {
	p.bus.log.Printf("Reply: %.80s", p.String())
	if err := p.bus.reply(p); err != nil {
		p.bus.log.Println(err)
	}
}

func Broadcast(p *Packet) {
	p.Broadcast()
}

func (p *Packet) Broadcast() {
	p.bus.log.Printf("Broadcast: %.80s", p.String())
	if err := p.bus.broadcast(p); err != nil {
		p.bus.log.Println(err)
	}
}
