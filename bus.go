package merle

import (
	"fmt"
	"regexp"
	"sync"
)

// Subcriber calls a callback when a packet received on the bus matches the
// matching criteria.  The subscriber uses regular expression matching to test
// each packet message.  See https://github.com/google/re2/wiki/Syntax for
// regular expression syntax.
type Subscriber struct {
	// Regular expression (re) for packet matching.  The re could be
	// an exact match ("CmdStart"), or something more exotic ("Cmd.*").
	Msg string
	// The callback to call when a packet message matches.  To drop a
	// packet, specify nil for the callback.
	Cb func(*Packet)
}

// Subscibers is a list of subscribers.  On packet receipt, the subscribers are
// processed in-order, and the first matching subscriber stops the processing.
type Subscribers []Subscriber

type sockets map[socketer]bool
type socketQ chan bool

type bus struct {
	thing *Thing
	// sockets
	sockLock sync.RWMutex
	sockets  sockets
	socketQ  socketQ
	// message subscribers
	subs    Subscribers
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
	// Queue any plugin attemps beyond socketsMax
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
	// add to front of array for highest priority
	b.subs = append([]Subscriber{{msg, f}}, b.subs...)
}

// Receive matches the packet against subscriber.  The first matching
// subscriber get the packet.  If no subscribers match, the packet is
// dropped.
func (b *bus) receive(p *Packet) error {
	msg := struct{ Msg string }{}
	p.Unmarshal(&msg)

	// TODO optimization: compile regexps

	for _, sub := range b.subs {
		matched, err := regexp.MatchString(sub.Msg, msg.Msg)
		if err != nil {
			return fmt.Errorf("Error compiling regexp \"%s\": %s", sub.Msg, err)
		}
		if matched {
			if sub.Cb != nil {
				b.thing.log.Printf("Received: %.80s", p.String())
				sub.Cb(p)
			}
			// first match wins
			return nil
		}
	}

	b.thing.log.Printf("Not handled: %.80s", p.String())

	return nil
}

// Reply sends the packet back to the source socket
func (b *bus) reply(p *Packet) error {
	if p.src == nil {
		return fmt.Errorf("Reply aborted; source is missing")
	}

	p.src.Send(p)

	return nil
}

// Broadcast sends the packet to each socket on the bus, expect to thexi
// originating socket
func (b *bus) broadcast(p *Packet) error {
	src := p.src

	b.sockLock.RLock()
	defer b.sockLock.RUnlock()

	if len(b.sockets) == 0 {
		return fmt.Errorf("Would broadcast: %.80s", p.String())
	}

	if len(b.sockets) == 1 && src != nil {
		if _, ok := b.sockets[src]; ok {
			return fmt.Errorf("Would broadcast: %.80s", p.String())
		}
	}

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then send on each connection

	for sock := range b.sockets {
		if sock == src {
			// don't send back to src
			continue
		}
		sock.Send(p)
	}

	return nil
}

// Send the packet to the destination socket
func (b *bus) send(p *Packet, dst socketer) error {
	dst.Send(p)
	return nil
}

func (b *bus) close() {
	b.sockLock.RLock()
	defer b.sockLock.RUnlock()

	for sock := range b.sockets {
		sock.Close()
	}
}
