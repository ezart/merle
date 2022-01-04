package merle

import (
	"reflect"
	"regexp"
	"sync"
	"log"
)

// Bus is a message bus.  Connect to the bus using Sockets.
type bus struct {
	log      *log.Logger
	// sockets
	sockLock sync.RWMutex
	sockets  map[ISocket] bool
	socketQ  chan bool
	// message subscribers
	subLock  sync.RWMutex
	subs     map[string][]func(*Packet)
}

func NewBus(log *log.Logger, socketsMax uint) *bus {
	return &bus{
		log: log,
		sockets: make(map[ISocket]bool),
		socketQ: make(chan bool, socketsMax),
		subs: make(map[string][]func(*Packet)),
	}
}

// Plug conection into bus
func (b *bus) plugin(s ISocket) {
	// Queue any plugin attemps beyond socketsMax
	b.socketQ <- true

	b.sockLock.Lock()
	defer b.sockLock.Unlock()
	b.sockets[s] = true
}

// Unplug conection from bus
func (b *bus) unplug(s ISocket) {
	b.sockLock.Lock()
	defer b.sockLock.Unlock()
	delete(b.sockets, s)

	<-b.socketQ
}

// Subscribe to message
func (b *bus) subscribe(msg string, f func(*Packet)) {
	b.subLock.Lock()
	defer b.subLock.Unlock()
	b.subs[msg] = append(b.subs[msg], f)
	b.log.Printf("Subscribed to \"%s\"", msg)
}

// Unsubscribe to message
func (b *bus) unsubscribe(msg string, f func(*Packet)) {
	b.subLock.Lock()
	defer b.subLock.Unlock()

	if _, ok := b.subs[msg]; !ok {
		return
	}

	for i, g := range b.subs[msg] {
		if reflect.ValueOf(g).Pointer() == reflect.ValueOf(f).Pointer() {
			b.log.Printf("Unsubscribed to \"%s\"", msg)
			b.subs[msg] = append(b.subs[msg][:i], b.subs[msg][i+1:]...)
			break
		}
	}

	if len(b.subs[msg]) == 0 {
		delete(b.subs, msg)
	}
}

func (b *bus) receive(p *Packet) {
	b.log.Printf("Received [%s]: %.80s", p.src.Name(), p.String())

	msg := struct {Msg string}{}
	p.Unmarshal(&msg)

	b.subLock.RLock()
	defer b.subLock.RUnlock()

	for key, funcs := range b.subs {
		matched, err := regexp.MatchString(key, msg.Msg)
		if err != nil {
			b.log.Printf("Error compiling regexp \"%s\": %s", key, err)
			continue
		}
		if matched {
			for _, f := range funcs {
				f(p)
			}
		}
	}
}

func (b *bus) reply(p *Packet) {
	if p.src == nil {
		return
	}

	b.log.Println("Reply", p.String())
	p.src.Send(p)
}

func (b *bus) broadcast(p *Packet) {
	src := p.src

	b.sockLock.RLock()
	defer b.sockLock.RUnlock()

	if len(b.sockets) == 0 {
		b.log.Printf("Would broadcast: %.80s", p.String())
		return
	}

	if len(b.sockets) == 1 && src != nil {
		if _, ok := b.sockets[src]; ok {
			b.log.Printf("Would broadcast: %.80s", p.String())
			return
		}
	}

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then send on each connection

	b.log.Printf("Broadcast: %.80s", p.String())
	for sock := range b.sockets {
		if sock == src {
			// don't send back to src
			continue
		}
		sock.Send(p)
	}
}

func (b *bus) send(p *Packet, sock ISocket) {
	b.log.Println("Send", p.String())
	sock.Send(p)
}

func (b *bus) close() {
	b.sockLock.RLock()
	defer b.sockLock.RUnlock()

	for sock := range b.sockets {
		sock.Close()
	}
}
















