package merle

import (
	"encoding/json"
	"log"
	"time"
	"net"
	"net/http"
	"sync"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Thing struct {
	Init func() error
	Run func()
	Home func(w http.ResponseWriter, r *http.Request)

	Status        string
	Id            string
	Model         string
	Name          string
	StartupTime   time.Time
	ConnsMax      int
	Shadow        bool

	things        map[string]*Thing

	sync.Mutex
	msgHandlers   map[string]func(*Packet)
	conns         map[*websocket.Conn]bool
	connQ         chan bool
	port          *port
	NewConnection chan *Thing

	// http servers
	sync.WaitGroup
	authUser      string
	portPublic    int
	portPrivate   int
	muxPublic     *mux.Router
	muxPrivate    *mux.Router
	httpPublic    *http.Server
	httpPrivate   *http.Server

	// tunnel to mother
	hubHost       string
	hubUser       string
	hubKey        string
}

func (t *Thing) connAdd(c *websocket.Conn) {
	t.Lock()
	defer t.Unlock()
	t.conns[c] = true
}

func (t *Thing) connDelete(c *websocket.Conn) {
	t.Lock()
	defer t.Unlock()
	delete(t.conns, c)
}

func (t *Thing) logPrefix() string {
	if t.Shadow {
		return "["+t.Id+","+t.Model+","+t.Name+"]"
	}
	return ""
}

func (t *Thing) identify(p *Packet) {
	resp := struct {
		Msg         string
		Status      string
		Id          string
		Model       string
		Name        string
		StartupTime time.Time
	}{
		Msg:         "identity",
		Status:      t.Status,
		Id:          t.Id,
		Model:       t.Model,
		Name:        t.Name,
		StartupTime: t.StartupTime,
	}
	p.Msg, _ = json.Marshal(&resp)
	t.Reply(p)
}

func (t *Thing) getThing(id string) *Thing {
	if thing, ok := t.things[id]; ok {
		return thing
	}
	return nil
}

func (t *Thing) receive(p *Packet) {
	msg := struct {
		Msg string
	}{}

	json.Unmarshal(p.Msg, &msg)

	f := t.msgHandlers[msg.Msg]
	if f == nil {
		log.Printf("%s Skipping msg; no handler: %.80s",
			t.logPrefix(), p.Msg)
	}

	f(p)
}

// Add a message handler
func (t *Thing) HandleMsg(msg string, f func(*Packet)) {
	if t.msgHandlers == nil {
		t.msgHandlers = make(map[string]func(*Packet))
	}
	t.msgHandlers[msg] = f
}

// Configure local http server
func (t *Thing) HttpConfig(authUser string, portPublic, portPrivate int) {
	t.authUser = authUser
	t.portPublic = portPublic
	t.portPrivate = portPrivate
}

// Start the Thing
func (t *Thing) Start() {
	t.conns = make(map[*websocket.Conn]bool)

	if t.ConnsMax == 0 {
		t.ConnsMax = 10
	}
	t.connQ = make(chan bool, t.ConnsMax)

	if t.Shadow {
		return
	}

	t.httpInit()

	if t.Init != nil {
		log.Println(t.logPrefix(), "Init...")
		if err := t.Init(); err != nil {
			log.Fatalln(t.logPrefix(), "Init failed:", err)
		}
	}

	t.HandleMsg("identify", t.identify)

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

// Sink sends Packet down towards bottom-most non-shadow Thing
func (t *Thing) Sink(p *Packet) {
	if !t.Shadow {
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

// Broadcast packet to all except self
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

func (t *Thing) changeStatus(child *Thing, status string) {
	/*
	child.Status = status

	spam := struct {
		Msg     string
		Id      string
		Status  string
	}{
		Msg:    "status",
		Id:     child.Id,
		Status: child.Status,
	}

	msg, _ := json.Marshal(&spam)
	t.broadcast(msg)
	*/
}

func (t *Thing) portRun(p *port) {
	/*
	var child *Thing

	resp, err := p.connect()
	if err != nil {
		goto disconnect
	}

	child = t.getThing(resp.Id)
	if child == nil {
		d = h.newDevice(resp.Id, resp.Model, resp.Name, resp.StartupTime)
		if d == nil {
			goto disconnect
		}
	} else {
		d.model = resp.Model
		d.name = resp.Name
		d.startupTime = resp.StartupTime
	}

	err = h.saveDevice(d)
	if err != nil {
		goto disconnect
	}

	h.changeStatus(d, "online")
	p.run(d)
	h.changeStatus(d, "offline")

disconnect:
	p.disconnect()
	*/
}

