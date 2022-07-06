// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"encoding/json"
)

// A Packet is the basic unit of communication in Merle.  Message Subscribers
// receive, process and optional forward a Packet.
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

// JSON marshal into Packet message
func (p *Packet) Marshal(msg interface{}) *Packet {
	p.msg, _ = json.Marshal(msg)
	return p
}

// JSON unmarshal from Packet message
func (p *Packet) Unmarshal(msg interface{}) {
	json.Unmarshal(p.msg, msg)
}

// String representation of Packet message
func (p *Packet) String() string {
	return string(p.msg)
}

func (p *Packet) Src() string {
	if p.src == nil {
		return "???"
	}
	return p.src.Src()
}

// Reply back to sender of Packet.
func (p *Packet) Reply() {
	p.bus.reply(p)
}

// Broadcast the Packet to all listeners except for the source of the Packet.
func (p *Packet) Broadcast() {
	p.bus.broadcast(p)
}

// Send Packet to destination
func (p *Packet) Send(dst string) {
	p.bus.send(p, dst)
}

func (p *Packet) IsThing() bool {
	return !p.bus.thing.isPrime
}

// Subscriber callback function to broadcast packet.  In this example, any
// packets received with message Alert are broadcast to all other listeners.
// Not applicable for CmdRun.
//
//	return merle.Subscribers{
//		{"Alert", merle.Broadcast},
//	}
//
func Broadcast(p *Packet) {
	msg := Msg{}
	p.Unmarshal(&msg)
	if msg.Msg == CmdRun {
		return
	}
	p.Broadcast()
}

func NoInit(p *Packet) {
}
// Subscriber callback function to run forever.  Only applicable for CmdRun.
// Use this callback when there is no other work to do in CmdRun than select{}.
//
//	return merle.Subscribers{
//		{CmdRun, merle.RunForver},
//	}
//
func RunForever(p *Packet) {
	msg := Msg{}
	p.Unmarshal(&msg)
	if msg.Msg != CmdRun {
		return
	}
	select {}
}

func ReplyStateEmpty(p *Packet) {
	msg := Msg{Msg: ReplyState}
	p.Marshal(&msg).Reply()
}

func ReplyGetState(p *Packet) {
	msg := Msg{Msg: GetState}
	p.Marshal(&msg).Reply()
}

func ReplyGetIdentity(p *Packet) {
	msg := Msg{Msg: GetIdentity}
	p.Marshal(&msg).Reply()
}
