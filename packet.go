// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"encoding/json"
)

// A Packet is the basic unit of communication in Merle.  A Packet contains a
// message.  Thing message Subscribers receive, process and optional forward
// Packets.
type Packet struct {
	// Bus the Packet lives on
	bus *bus
	// Source socket on bus; the Packet source
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

// Src is the Packet's source Thing Id.  If the Packet originated internally
// (p.src == nil), then use "SYSTEM" for Id.
func (p *Packet) Src() string {
	if p.src == nil {
		return "SYSTEM"
	}
	return p.src.Src()
}

// Reply back to sender of Packet.  Do not hold locks when calling Reply().
func (p *Packet) Reply() {
	p.bus.reply(p)
}

// Broadcast the Packet to everyone else on the bus.  Do not hold locks when
// call Broadcast().
func (p *Packet) Broadcast() {
	p.bus.broadcast(p)
}

// Send Packet to destination
// TODO: Use restrictions?  Only to be called from bridge, or could be called
// TODO: from child to talk to another child, over a bridge?
func (p *Packet) Send(dst string) {
	p.bus.send(p, dst)
}

// Test if this is the real Thing or Thing Prime.
//
// If p.IsThing() is not true, then we're on Thing Prime and should not access
// device I/O and only update Thing's software state.  If p.IsThing() is true,
// then this is the real Thing and we can access device I/O.
func (p *Packet) IsThing() bool {
	return !p.bus.thing.isPrime
}

// Subscriber callback function to broadcast Packet.  In this example, any
// Packets received with message Alert are broadcast to all other listeners:
//
//	return merle.Subscribers{
//		...
//		"Alert", merle.Broadcast,
//	}
//
func Broadcast(p *Packet) {
	p.Broadcast()
}

// Subscriber callback function to skip CmdInit.  In this example, CmdInit is
// skipped:
//
//	return merle.Subscribers{
//		merle.CmdInit: merle.NoInit,
//		...
//	}
//
func NoInit(p *Packet) {
}

// Subscriber callback function to run forever.  Only applicable for CmdRun.
// Use this callback when there is no other work to do in CmdRun than select{}.
//
//	return merle.Subscribers{
//		...
//		merle.CmdRun: merle.RunForver,
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

// Subscriber callback function to return empty state in response to GetState.
// Example:
//
//	return merle.Subscribers{
//		...
//		merle.GetState: merle.ReplyStateEmpty,
//	}
//
func ReplyStateEmpty(p *Packet) {
	msg := Msg{Msg: ReplyState}
	p.Marshal(&msg).Reply()
}

// Subscriber callback function to GetState
func ReplyGetState(p *Packet) {
	msg := Msg{Msg: GetState}
	p.Marshal(&msg).Reply()
}

// Subscriber callback function to GetIdentity.  Example of chaining the event
// status change notification to send a GetIdentity request:
//
//	return merle.Subscribers{
//		...
//		merle.EventStatus: merle.ReplyGetIdentity,
//		merle.ReplyIdentity: t.identity,
//	}
//
func ReplyGetIdentity(p *Packet) {
	msg := Msg{Msg: GetIdentity}
	p.Marshal(&msg).Reply()
}
