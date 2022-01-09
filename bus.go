package merle

import (
	"fmt"
	"log"
	"regexp"
	"sync"
)

type subscribers map[string][]func(*Packet)
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
	blocked    subscribers
	allowed    subscribers
}

func newBus(log *log.Logger, socketsMax uint, subs Subscribers) *bus {
	b := &bus{
		log:     log,
		sockets: make(sockets),
		socketQ: make(socketQ, socketsMax),
		blocked: make(subscribers),
		allowed: make(subscribers),
	}

	return b.sort(subs)
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

func (b *bus) sort(subs Subscribers) *bus {
	for msg, f := range subs {
		b.subscribe(msg, f)
	}
	return b
}

// Subscribe to message
func (b *bus) subscribe(msg string, f func(*Packet)) {
	if msg == "" {
		return
	}

	block := false
	if msg[0:1] == "-" {
		block = true
		msg = msg[1:]
	}

	b.subLock.Lock()
	defer b.subLock.Unlock()
	if block {
		b.blocked[msg] = append(b.blocked[msg], nil)
	} else {
		b.allowed[msg] = append(b.allowed[msg], f)
	}
}

func (b *bus) receive(p *Packet) error {
	msg := struct{ Msg string }{}
	p.Unmarshal(&msg)

	b.subLock.RLock()
	defer b.subLock.RUnlock()

	for key, _ := range b.blocked {
		matched, err := regexp.MatchString(key, msg.Msg)
		if err != nil {
			return fmt.Errorf("Error compiling regexp \"%s\": %s", key, err)
		}
		if matched {
			return nil
		}
	}

	b.log.Printf("Receive: %.80s", p.String())

	for key, funcs := range b.allowed {
		matched, err := regexp.MatchString(key, msg.Msg)
		if err != nil {
			return fmt.Errorf("Error compiling regexp \"%s\": %s", key, err)
		}
		if matched {
			for _, f := range funcs {
				f(p)
			}
		}
	}

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
