package merle

import (
	"encoding/json"
	"log"
	"time"
	"net"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

type port struct {
	ws *websocket.Conn
}

type Thing struct {
	Init func() error
	Run func()
	Home func(w http.ResponseWriter, r *http.Request)

	Status        string
	Id            string
	Model         string
	Name          string
	StartupTime   time.Time

	sync.Mutex
	handlers      map[string]func(*Packet)
	conns         map[*websocket.Conn]bool
	port          *port
	inHub         bool

	// http servers
	sync.WaitGroup
	authUser      string
	portPublic    int
	portPrivate   int
	httpPublic    *http.Server
	httpPrivate   *http.Server

	// tunnel to hub
	hubHost       string
	hubUser       string
	hubKey        string
}

func (d *Thing) connAdd(c *websocket.Conn) {
	d.Lock()
	defer d.Unlock()
	d.conns[c] = true
}

func (d *Thing) connDelete(c *websocket.Conn) {
	d.Lock()
	defer d.Unlock()
	delete(d.conns, c)
}

func (t *Thing) logPrefix() string {
	if t.inHub {
		return "["+t.Id+","+t.Model+","+t.Name+"]"
	}
	return ""
}

func (t *Thing) identify(p *Packet) {
	resp := struct {
		Type        string
		Status      string
		Id          string
		Model       string
		Name        string
		StartupTime time.Time
	}{
		Type:        "identity",
		Status:      t.Status,
		Id:          t.Id,
		Model:       t.Model,
		Name:        t.Name,
		StartupTime: t.StartupTime,
	}
	p.Msg, _ = json.Marshal(&resp)
	t.Reply(p)
}

func (t *Thing) receive(p *Packet) {
	msg := struct {
		Type string
	}{}

	json.Unmarshal(p.Msg, &msg)

	f := t.handlers[msg.Type]
	if f == nil {
		log.Printf("%s Skipping msg; no handler: %.80s",
			t.logPrefix(), p.Msg)
	}

	f(p)
}

// Add a message handler
func (t *Thing) AddHandler(msgType string, f func(*Packet)) {
	if t.handlers == nil {
		t.handlers = make(map[string]func(*Packet))
	}
	t.handlers[msgType] = f
}

// Configure local http server
func (t *Thing) HttpConfig(authUser string, portPublic, portPrivate int) {
	t.authUser = authUser
	t.portPublic = portPublic
	t.portPrivate = portPrivate
}

// Start the thing
func (t *Thing) Start() {
	t.conns = make(map[*websocket.Conn]bool)

	if t.Init != nil {
		log.Println(t.logPrefix(), "Init...")
		if err := t.Init(); err != nil {
			log.Fatalln(t.logPrefix(), "Init failed:", err)
		}
	}

	t.AddHandler("identify", t.identify)

	t.tunnelCreate()

	t.httpStart()

	if t.Run != nil {
		log.Println(t.logPrefix(), "Run...")
		t.Run()
	}

	t.httpStop()

	log.Fatalln(t.logPrefix(), "Run() didn't run forever")
}

// Reply sends Packet back to originator
func (t *Thing) Reply(p *Packet) {
	log.Printf("%s Reply: %.80s", t.logPrefix(), p.Msg)

	t.Lock()
	defer t.Unlock()

	err := p.writeMessage()
	if err != nil {
		log.Println(t.logPrefix(), "Reply error:", err)
	}
}

// Sink sends Packet down towards bottom-most thing.
func (t *Thing) Sink(p *Packet) {
	if !t.inHub {
		return
	}

	src := p.conn

	t.Lock()
	defer func() {
		p.conn = src
		t.Unlock()
	}()

	if t.port == nil {
		log.Printf("%s Sink error: not running on port: %.80s",
			t.logPrefix(), p.Msg)
		return
	}

	if src == t.port.ws {
		log.Printf("%s Sink reject: message came in on port: %.80s",
			t.logPrefix(), p.Msg)
		return
	}

	log.Printf("%s Sink: %.80s", t.logPrefix(), p.Msg)

	p.conn = t.port.ws

	err := p.writeMessage()
	if err != nil {
		log.Println(t.logPrefix(), "Sink error:", err)
	}
}

// Broadcast packet to all except self.
func (t *Thing) Broadcast(p *Packet) {
	src := p.conn

	t.Lock()
	defer func() {
		p.conn = src
		t.Unlock()
	}()

	switch len(t.conns) {
	case 0:
		log.Printf("%s Would broadcast: %.80s", t.logPrefix(), p.Msg)
		return
	case 1:
		if _, ok := t.conns[src]; ok {
			log.Printf("%s Would broadcast: %.80s",
				t.logPrefix(), p.Msg)
			return
		}
	}

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then send on each connection

	for c := range t.conns {
		if c == src {
			// skip self
			log.Printf("%s Skipping broadcast: %.80s",
				t.logPrefix(), p.Msg)
			continue
		}
		p.conn = c
		log.Printf("%s Broadcast: %.80s", t.logPrefix(), p.Msg)
		p.writeMessage()
	}
}

func (t *Thing) HomeParams(r *http.Request) interface{} {
	scheme := "wss:\\"
	if r.TLS == nil {
		scheme = "ws:\\"
	}

	return struct {
		Scheme string
		Host   string
		Status string
		Id     string
		Model  string
		Name   string
	}{
		Scheme: scheme,
		Host:   r.Host,
		Status: t.Status,
		Id:     t.Id,
		Model:  t.Model,
		Name:   t.Name,
	}
}

// DefaultId returns a default ID based on the device's MAC address
func DefaultId_() string {

	// Use the MAC address of the first non-lo interface
	// as the default device ID

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Name != "lo" {
				return iface.HardwareAddr.String()
			}
		}
	}

	return ""
}
