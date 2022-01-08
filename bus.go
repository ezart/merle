package merle

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"
)

type sockets map[socketer]bool
type socketQ chan bool

// Bus is a message bus.  Connect to the bus using Sockets.
type bus struct {
	// sockets
	sockLock sync.RWMutex
	sockets  sockets
	socketQ  socketQ
	// message subscribers
	subLock sync.RWMutex
	subs    Subscribers
}

func newBus(socketsMax uint, subs Subscribers) *bus {
	return &bus{
		sockets: make(sockets),
		socketQ: make(socketQ, socketsMax),
		subs:    subs,
	}
}

// Plug conection into bus
func (b *bus) plugin(s socketer) {
	// Queue any plugin attemps beyond socketsMax
	b.socketQ <- true

	b.sockLock.Lock()
	b.sockets[s] = true
	b.sockLock.Unlock()
}

// Unplug conection from bus
func (b *bus) unplug(s socketer) {
	b.sockLock.Lock()
	delete(b.sockets, s)
	b.sockLock.Unlock()

	<-b.socketQ
}

// Subscribe to message
func (b *bus) subscribe(msg string, f func(*Packet)) {
	b.subLock.Lock()
	defer b.subLock.Unlock()
	b.subs[msg] = append(b.subs[msg], f)
}

// Unsubscribe to message
func (b *bus) unsubscribe(msg string, f func(*Packet)) error {
	b.subLock.Lock()
	defer b.subLock.Unlock()

	if _, ok := b.subs[msg]; !ok {
		return fmt.Errorf("Not subscribed to \"%s\"", msg)
	}

	found := false
	for i, g := range b.subs[msg] {
		if reflect.ValueOf(g).Pointer() == reflect.ValueOf(f).Pointer() {
			found = true
			b.subs[msg] = append(b.subs[msg][:i], b.subs[msg][i+1:]...)
			break
		}
	}

	if !found {
		return fmt.Errorf("Not subscribed to \"%s\" for function", msg)
	}

	if len(b.subs[msg]) == 0 {
		delete(b.subs, msg)
	}

	return nil
}

func (b *bus) receive(p *Packet) error {
	msg := struct{ Msg string }{}
	p.Unmarshal(&msg)

	b.subLock.RLock()
	defer b.subLock.RUnlock()

	for key, funcs := range b.subs {
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

func (b *bus) close() {
	b.sockLock.RLock()
	defer b.sockLock.RUnlock()

	for sock := range b.sockets {
		sock.Close()
	}
}
