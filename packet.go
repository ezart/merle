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
	src ISocket
	msg []byte
}

func newPacket(bus *bus, src ISocket, msg interface{}) *Packet {
	p := &Packet{bus: bus, src: src}
	p.msg, _ = json.Marshal(msg)
	return p
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

func (p *Packet) Reply() {
	p.bus.reply(p)
}

func (p *Packet) Broadcast() {
	p.bus.broadcast(p)
}

func (p *Packet) Send(sock ISocket) {
	p.bus.send(p, sock)
}

func (p *Packet) Multicast(socks...ISocket) {
	for _, sock := range socks {
		sock.Send(p)
	}
}
