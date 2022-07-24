// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import "sync"

// Subscribers is a map of message subscribers, keyed by Msg.  On Packet
// receipt, the Packet Msg is used to lookup a subscriber.  If a match,
// the subscriber handler is called to process the Packet.
//
// Here's an example Subscribers() list:
//
// func (t *thing) Subscribers() merle.Subscribers {
//	return merle.Subscribers{
//		merle.CmdInit:     t.init,
//		merle.CmdRun:      t.run,
//		merle.GetState:    t.getState,
//		merle.EventStatus: nil,
//		"SetPoint":        t.setPoint,
//	}
//
// A subscriber handler is a function that takes a Packet pointer as it's only
// argument.  An example handler for the "SetPoint" Msg above:
//
// func (t *thing) setPoint(p *merle.Packet) {
//	// do something with Packet p
// }
//
// If the handler is nil, a Packet will be dropped silently.
//
// If the key "default" exists, then the default handler is called for any
// non-matching Packets.  Here's an example BridgeSuscribers() that silently
// drops all packets except CAN messages:
//
// func (b *bridge) BridgeSubscribers() merle.Subscribers {
// 	return merle.Subscribers{
// 		"CAN":     merle.Broadcast, // broadcast CAN msgs to everyone
// 		"default": nil,             // drop everything else silently
// 	}
// }
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

// Receive matches the packet against subscribers and calls the matching
// subscriber handler.  If no subscribers match the received message, the
// "default" subscriber matches.  If still no matches, the packet is (silently)
// dropped.
func (b *bus) receive(p *Packet) {
	var msg Msg

	p.Unmarshal(&msg)

	f, match := b.subs[msg.Msg]
	if match {
		if f != nil {
			b.thing.log.printf("Received [%s]: %.80s", p.Src(),
				p.String())
			f(p)
		}
	} else {
		f, match = b.subs["default"]
		if match {
			if f != nil {
				b.thing.log.printf("Received [%s] by default: %.80s",
					p.Src(), p.String())
				f(p)
			}
		} else {
			b.thing.log.printf("Not handled [%s]: %.80s", p.Src(),
				p.String())
		}
	}

	// Receiving ReplyState is a special case.  The socket is disabled for
	// broadcasts until ReplyState is received.

	if msg.Msg == ReplyState {
		p.src.SetFlags(p.src.Flags() | sock_flag_bcast)
		b.thing.log.println("GOT REPLY STATE bcast set", p.src.Name())
	}
}

// Reply sends the packet back to the source socket
func (b *bus) reply(p *Packet) {
	if p.src == nil {
		b.thing.log.println("REPLY ABORTED; source is missing")
		return
	}

	msg := Msg{}
	p.Unmarshal(&msg)

	b.thing.log.printf("Reply: %.80s", p.String())
	p.src.Send(p)

	// Sending ReplyState is a special case.  The socket is disabled for
	// broadcasts until ReplyState is sent.  This ensures other end doesn't
	// receive unsolicited broadcast messages before ReplyState.

	if msg.Msg == ReplyState {
		p.src.SetFlags(p.src.Flags() | sock_flag_bcast)
		b.thing.log.println("SENDING REPLY STATE bcast set", p.src.Name())
	}
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
			b.thing.log.println("SKIPPING broadcast to SELF:", sock.Name())
			continue
		}
		if sock.Flags()&sock_flag_bcast == 0 {
			// Socket not ready for broadcasts.  Once a ReplyState
			// message has been processed, the socket will be
			// enabled for broadcasts.
			b.thing.log.println("SKIPPING BCAST NOT SET:", sock.Name())
			continue
		}
		if sent == 0 {
			b.thing.log.printf("Broadcast: %.80s", p.String())
			sent++
		}
		sock.Send(p)
	}

	if sent == 0 {
		b.thing.log.printf("Would Broadcast: %.80s", p.String())
	}
}

func (b *bus) send(p *Packet, dst string) {
	sent := false

	b.sockLock.RLock()
	defer b.sockLock.RUnlock()

	for sock := range b.sockets {
		if sock.Src() == dst {
			b.thing.log.printf("Send to [%s]: %.80s", dst, p.String())
			sock.Send(p)
			sent = true
			break
		}
	}

	if !sent {
		b.thing.log.printf("Destination [%s] unknown: %.80s", dst, p.String())
	}
}

func (b *bus) close() {
	b.thing.log.println("CLOSING BUS")
	b.sockLock.Lock()
	defer b.sockLock.Unlock()

	for sock := range b.sockets {
		sock.Close()
		delete(b.sockets, sock)
	}
}
