// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

// Socket flags
const (
	bcast uint32 = 1 << iota
)

// Socketer is an interface to a socket.  A socket plugs into a bus.
type socketer interface {
	// Send the packet on bus the socket is connected to
	Send(*Packet) error
	// Close the socket
	Close()
	// Name of the socket
	Name() string
	Flags() uint32
	SetFlags(uint32)
	// Packet src thing ID
	Id() string
}

// Wire socket
type wireSocket struct {
	name     string
	flags    uint32
	bus      *bus
	opposite *wireSocket
}

func newWireSocket(name string, bus *bus, opposite *wireSocket) *wireSocket {
	return &wireSocket{name: name, flags: bcast, bus: bus, opposite: opposite}
}

func (s *wireSocket) Send(p *Packet) error {
	s.bus.receive(p.clone(s.bus, s.opposite))
	return nil
}

func (s *wireSocket) Close() {
}

func (s *wireSocket) Name() string {
	return s.name
}

func (s *wireSocket) Flags() uint32 {
	return s.flags
}

func (s *wireSocket) SetFlags(flags uint32) {
	s.flags = flags
}

func (s *wireSocket) Id() string {
	return s.bus.thing.id
}
