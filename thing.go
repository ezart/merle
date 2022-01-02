package merle

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"regexp"
	"reflect"
	"sync"
	"time"
)

type Thing struct {
	Init func() error
	Run  func()
	Home func(w http.ResponseWriter, r *http.Request)
	Connect func(*Thing)

	status      string
	id          string
	model       string
	name        string
	startupTime time.Time
	shadow      bool
	connsMax    int
	cfgFile     string
	demoMode    bool

	// children
	stork    func(string, string, string) *Thing
	children map[string]*Thing

	// ws connections
	connLock      sync.RWMutex
	conns         map[*websocket.Conn]bool
	connQ         chan bool

	// msg subscribers
	subLock       sync.RWMutex
	subscribers   map[string][]func(*Packet)

	// http servers
	sync.WaitGroup
	authUser    string
	portPublic  int
	portPrivate int
	muxPublic   *mux.Router
	muxPrivate  *mux.Router
	httpPublic  *http.Server
	httpPrivate *http.Server

	// mother
	motherHost string
	motherUser string
	motherKey  string
	motherPortPrivate int
}

func (t *Thing) Id() string {
	return t.id
}

func (t *Thing) Status() string {
	return t.status
}

func (t *Thing) SetStork(f func(string, string, string) *Thing) {
	t.stork = f
}

func (t *Thing) SetConfigFile(cfgFile string) {
	t.cfgFile = cfgFile
}

func (t *Thing) ConfigFile() string {
	return t.cfgFile
}

func (t *Thing) SetDemoMode(demoMode bool) {
	t.demoMode = demoMode
}

func (t *Thing) DemoMode() bool {
	return t.demoMode
}

func (t *Thing) InitThing(id, model, name string) *Thing {
	if model == "" {
		log.Println("Thing Model is missing")
		return nil
	}
	if name == "" {
		log.Println("Thing Name is missing")
		return nil
	}
	if id == "" {
		id = defaultId()
		log.Println("Thing ID defaulting to", id)
	}

	t.id = id
	t.model = model
	t.name = name
	t.status = "online"
	t.startupTime = time.Now()

	t.conns = make(map[*websocket.Conn]bool)

	if t.connsMax == 0 {
		t.connsMax = 10
	}
	t.connQ = make(chan bool, t.connsMax)

	t.stork = func(string, string, string) *Thing {
		log.Println(t.logPrefix(), "Need to set stork")
		return nil
	}
	t.children = make(map[string]*Thing)

	t.Subscribe("GetIdentity", t.getIdentity)

	return t
}

func (t *Thing) connAdd(c *websocket.Conn) {
	t.connLock.Lock()
	defer t.connLock.Unlock()
	t.conns[c] = true
}

func (t *Thing) connDel(c *websocket.Conn) {
	t.connLock.Lock()
	defer t.connLock.Unlock()
	delete(t.conns, c)
}

func (t *Thing) logPrefix() string {
	if t.shadow {
		return "[" + t.id + "," + t.model + "," + t.name + "] "
	}
	return ""
}

type msgIdentity struct {
	Msg         string
	Status      string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}

func (t *Thing) getIdentity(p *Packet) {
	resp := msgIdentity {
		Msg:         "ReplyIdentity",
		Status:      t.status,
		Id:          t.id,
		Model:       t.model,
		Name:        t.name,
		StartupTime: t.startupTime,
	}
	t.Reply(p.Marshal(&resp))
}

type msgThing struct {
	Id string
	Model string
	Name string
	Status string
}

type msgThings struct {
	Msg string
	Things []msgThing
}

func (t *Thing) getThings(p *Packet) {
	resp := msgThings{
		Msg: "ReplyThings",
	}
	for _, child := range t.children {
		resp.Things = append(resp.Things, msgThing{child.id,
			child.model, child.name, child.status})
	}
	t.Reply(p.Marshal(&resp))
}

func (t *Thing) GetChild(id string) *Thing {
	if thing, ok := t.children[id]; ok {
		return thing
	}
	return nil
}

func (t *Thing) receive(p *Packet) {
	log.Printf("%sReceived: %.80s", t.logPrefix(), p.String())

	msg := struct {
		Msg string
	}{}

	p.Unmarshal(&msg)

	t.subLock.RLock()
	defer t.subLock.RUnlock()

	for key, subscribers := range t.subscribers {
		matched, err := regexp.MatchString(key, msg.Msg)
		if err != nil {
			log.Printf("%sError compiling regexp \"%s\": %s", t.logPrefix(), key, err)
			continue
		}
		if matched {
			for _, f := range subscribers {
				f(p)
			}
		}
	}
}

// Subscribe to message
func (t *Thing) Subscribe(msg string, f func(*Packet)) {
	t.subLock.Lock()
	defer t.subLock.Unlock()

	if t.subscribers == nil {
		t.subscribers = make(map[string][]func(*Packet))
	}
	t.subscribers[msg] = append(t.subscribers[msg], f)

	log.Printf("%sSubscribed to \"%s\"", t.logPrefix(), msg)
}

// Unsubscribe to message
func (t *Thing) Unsubscribe(msg string, f func(*Packet)) {
	t.subLock.Lock()
	defer t.subLock.Unlock()

	if t.subscribers == nil {
		return
	}

	if _, ok := t.subscribers[msg]; !ok {
		return
	}

	for i, g := range t.subscribers[msg] {
		if reflect.ValueOf(g).Pointer() == reflect.ValueOf(f).Pointer() {
			log.Printf("%sUnsubscribed to \"%s\"", t.logPrefix(), msg)
			t.subscribers[msg] = append(t.subscribers[msg][:i],
				t.subscribers[msg][i+1:]...)
			break
		}
	}

	if len(t.subscribers[msg]) == 0 {
		delete(t.subscribers, msg)
	}
}

// Configure local http server
func (t *Thing) HttpConfig(authUser string, portPublic, portPrivate int) {
	t.authUser = authUser
	t.portPublic = portPublic
	t.portPrivate = portPrivate
}

// Start the Thing
func (t *Thing) Start() {
	if t.demoMode {
		log.Println(t.logPrefix(), "Demo mode ENABLED")
	}

	if t.shadow {
		return
	}

	t.httpInit()

	if t.Init != nil {
		log.Println(t.logPrefix(), "Init...")
		if err := t.Init(); err != nil {
			log.Fatalln(t.logPrefix(), "Init failed:", err)
		}
	}

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
	t.connLock.RLock()
	defer t.connLock.RUnlock()

	log.Printf("%sReply: %.80s", t.logPrefix(), p.String())
	err := p.write()
	if err != nil {
		log.Println(t.logPrefix(), "Reply error:", err)
	}
}

// Broadcast packet to all except packet source
func (t *Thing) Broadcast(p *Packet) {
	src := p.conn

	t.connLock.RLock()
	defer func() {
		p.conn = src
		t.connLock.RUnlock()
	}()

	switch len(t.conns) {
	case 0:
		log.Printf("%sWould broadcast: %.80s", t.logPrefix(), p.String())
		return
	case 1:
		if _, ok := t.conns[src]; ok {
			log.Printf("%sWould broadcast: %.80s",
				t.logPrefix(), p.String())
			return
		}
	}

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then send on each connection

	log.Printf("%sBroadcast: %.80s", t.logPrefix(), p.String())
	for c := range t.conns {
		if c == src {
			// don't send back to src
			continue
		}
		p.conn = c
		p.write()
	}
}

func (t *Thing) HomeParams(r *http.Request, extra interface{}) interface{} {
	scheme := "wss://"
	if r.TLS == nil {
		scheme = "ws://"
	}

	return struct {
		Scheme string
		Host   string
		Status string
		Id     string
		Model  string
		Name   string
		Extra  interface{}
	}{
		Scheme: scheme,
		Host:   r.Host,
		Status: t.status,
		Id:     t.id,
		Model:  t.model,
		Name:   t.name,
		Extra:  extra,
	}
}

// DefaultId returns a default ID based on the device's MAC address
func defaultId() string {

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

type SpamStatus struct {
	Msg     string
	Id      string
	Model   string
	Name    string
	Status  string
}

func (t *Thing) changeStatus(child *Thing, status string) {
	child.status = status

	spam := SpamStatus{
		Msg:    "SpamStatus",
		Id:     child.id,
		Model:  child.model,
		Name:   child.name,
		Status: child.status,
	}
	p := NewPacket(&spam)
	t.Broadcast(p)

	if t.Connect != nil {
		t.Connect(child)
	}
}

func (t *Thing) portRun(p *port, match string) {
	var child *Thing

	resp, err := p.connect()
	if err != nil {
		goto disconnect
	}

	// TODO disconnect if resp doesn't match filter

	if t.id == resp.Id {
		log.Println(t.logPrefix(), "Sorry, you can't be your own Mother")
		goto disconnect
	}

	child = t.GetChild(resp.Id)

	if child == nil {
		child = t.stork(resp.Id, resp.Model, resp.Name)
		if child == nil {
			log.Println(t.logPrefix(), "Model", resp.Model, "unknown")
			goto disconnect
		}
		child.shadow = true
		t.children[resp.Id] = child
	} else {
		if child.model != resp.Model {
			log.Println(t.logPrefix(), "Model mismatch")
			goto disconnect
		}
		if child.name != resp.Name {
			log.Println(t.logPrefix(), "Name mismatch")
			goto disconnect
		}
	}

	child.startupTime = resp.StartupTime

	t.changeStatus(child, "online")
	p.run(child)
	t.changeStatus(child, "offline")

   disconnect:
	p.disconnect()
}
