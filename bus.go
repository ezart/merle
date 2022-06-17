// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"sync"
)

// Subscibers is a map of message subscribers, keyed by the message.  On packet
// receipt, the packet message is used to lookup a subsciber.  The subscriber
// callback is called to handle the packet.
type Subscribers map[string]func(*Packet)

type sockets map[socketer]bool
type socketQ chan bool

type bus struct {
	thing *Thing
	// sockets
	sockLock sync.RWMutex
	sockets  sockets
	socketQ  socketQ
	// message subscribers
	subs Subscribers
}

func newBus(thing *Thing, socketsMax uint, subs Subscribers) *bus {
	return &bus{
		thing:   thing,
		sockets: make(sockets),
		socketQ: make(socketQ, socketsMax),
		subs:    subs,
	}
}

// Plug a socket into the bus
func (b *bus) plugin(s socketer) {
	// Queue any plugin attempts beyond socketsMax
	b.socketQ <- true

	b.sockLock.Lock()
	b.sockets[s] = true
	b.sockLock.Unlock()
}

// Unplug a socket from the bus
func (b *bus) unplug(s socketer) {
	b.sockLock.Lock()
	delete(b.sockets, s)
	b.sockLock.Unlock()

	<-b.socketQ
}

// Subscribe to message
func (b *bus) subscribe(msg string, f func(*Packet)) {
	b.subs[msg] = f
}

// Receive matches the packet against subscribers.  If no subscribers match the
// received message, the "default" subscriber matches.  If still no matches,
// the packet is (silently) dropped.
func (b *bus) receive(p *Packet) {
	var msg Msg

	p.Unmarshal(&msg)

	f, match := b.subs[msg.Msg]
	if match {
		if f != nil {
			b.thing.log.Printf("Received: %.80s", p.String())
			f(p)
		}
		return
	}

	f, match = b.subs["default"]
	if match {
		if f != nil {
			b.thing.log.Printf("Received by default: %.80s", p.String())
			f(p)
		}
		return
	}

	b.thing.log.Printf("Not handled: %.80s", p.String())
}

// Reply sends the packet back to the source socket
func (b *bus) reply(p *Packet) {
	if p.src == nil {
		b.thing.log.Println("REPLY ABORTED; source is missing")
		return
	}

	msg := struct{ Msg string }{}
	p.Unmarshal(&msg)

	// Sending ReplyState is a special case.  The socket is disabled for
	// broadcasts until ReplyState is sent.  This ensures other end doesn't
	// receive unsolicited broadcast messages before ReplyState.

	if msg.Msg == ReplyState {
		p.src.SetFlags(p.src.Flags() | bcast)
	}

	b.thing.log.Printf("Reply: %.80s", p.String())
	p.src.Send(p)
}

// Broadcast sends the packet to each socket on the bus, expect to the
// originating socket
func (b *bus) broadcast(p *Packet) {
	sent := 0
	src := p.src

	b.sockLock.RLock()
	defer b.sockLock.RUnlock()

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then send on each connection

	for sock := range b.sockets {
		if sock == src {
			// don't send back to src
			b.thing.log.Println("SKIPPING SELF:", sock.Name())
			continue
		}
		if sock.Flags()&bcast == 0 {
			// Socket not ready for broadcasts.  Once a ReplyState
			// message has been sent, the socket will be enabled
			// for broadcasts.
			b.thing.log.Println("SKIPPING BCAST NOT SET:", sock.Name())
			continue
		}
		if sent == 0 {
			b.thing.log.Printf("Broadcast: %.80s", p.String())
			sent++
		}
		sock.Send(p)
	}

	if sent == 0 {
		b.thing.log.Printf("Would Broadcast: %.80s", p.String())
	}
}

func (b *bus) close() {
	b.sockLock.RLock()
	defer b.sockLock.RUnlock()

	for sock := range b.sockets {
		sock.Close()
	}
}
