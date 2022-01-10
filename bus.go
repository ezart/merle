package merle

import (
	"fmt"
	"log"
	"regexp"
	"sync"
)

type Subscriber struct {
	Msg string
	Cb func(*Packet)
}

type Subscribers []Subscriber

type sockets map[socketer]bool
type socketQ chan bool

type bus struct {
	log        *log.Logger
	// sockets
	sockLock   sync.RWMutex
	sockets    sockets
	socketQ    socketQ
	// message subscribers
	subLock    sync.RWMutex
	subs       Subscribers
}

func newBus(log *log.Logger, socketsMax uint, subs Subscribers) *bus {
	return &bus{
		log:     log,
		sockets: make(sockets),
		socketQ: make(socketQ, socketsMax),
		subs:    subs,
	}
}

func (b *bus) plugin(s socketer) {
	// Queue any plugin attemps beyond socketsMax
	b.socketQ <- true

	b.sockLock.Lock()
	b.sockets[s] = true
	b.sockLock.Unlock()
}

func (b *bus) unplug(s socketer) {
	b.sockLock.Lock()
	delete(b.sockets, s)
	b.sockLock.Unlock()

	<-b.socketQ
}

// Subscribe to message
func (b *bus) subscribe(msg string, f func(*Packet)) {
	b.subLock.Lock()
	// add to front of array for highest priority
	b.subs = append([]Subscriber{{msg, f}}, b.subs...)
	b.subLock.Unlock()
}

func (b *bus) receive(p *Packet) error {
	msg := struct{ Msg string }{}
	p.Unmarshal(&msg)

	b.subLock.RLock()
	defer b.subLock.RUnlock()

	// TODO optimization: compile regexps

	for _, sub := range b.subs {
		matched, err := regexp.MatchString(sub.Msg, msg.Msg)
		if err != nil {
			return fmt.Errorf("Error compiling regexp \"%s\": %s", sub.Msg, err)
		}
		if matched {
			if sub.Cb != nil {
				b.log.Printf("Received: %.80s", p.String())
				sub.Cb(p)
			}
			return nil
		}
	}

	b.log.Printf("Not handled: %.80s", p.String())

	return nil
}

func (b *bus) reply(p *Packet) error {
	if p.src == nil {
		return fmt.Errorf("Reply aborted; source is missing")
	}

	p.src.Send(p)

	return nil
}

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
