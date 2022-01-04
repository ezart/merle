package merle

import (
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type Thing struct {
	Init    func() error
	Run     func()
	Home    func(w http.ResponseWriter, r *http.Request)

	status      string
	id          string
	model       string
	name        string
	startupTime time.Time

	cfgFile  string
	demoMode bool
	log      *log.Logger
	inited   bool

	// message bus
	bus *bus

	// children
	stork       func(string, string, string) *Thing
	children    map[string]*Thing
	childStatus func(*Thing)

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
	motherHost        string
	motherUser        string
	motherKey         string
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

func (t *Thing) logPrefix() string {
	return "[" + t.id + "," + t.model + "," + t.name + "] "
}

func (t *Thing) InitThing(id, model, name string) *Thing {
	if t.inited {
		t.log.Println(t.logPrefix(), "Already inited!")
		return nil
	}

	if id == "" {
		id = defaultId()
	}

	t.id = id
	t.model = model
	t.name = name

	t.log = log.New(os.Stderr, t.logPrefix(), 0)

	if model == "" {
		t.log.Println("Thing Model is missing")
		return nil
	}
	if name == "" {
		t.log.Println("Thing Name is missing")
		return nil
	}

	t.status = "online"
	t.startupTime = time.Now()

	// TODO pass in connMax from cfg?
	t.bus = NewBus(10)

	t.stork = func(string, string, string) *Thing {
		t.log.Println("Need to set stork")
		return nil
	}
	t.children = make(map[string]*Thing)

	t.Subscribe("GetIdentity", t.getIdentity)

	t.inited = true

	return t
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
	resp := msgIdentity{
		Msg:         "ReplyIdentity",
		Status:      t.status,
		Id:          t.id,
		Model:       t.model,
		Name:        t.name,
		StartupTime: t.startupTime,
	}
	t.Reply(p.Marshal(&resp))
}

type msgChild struct {
	Id     string
	Model  string
	Name   string
	Status string
}

type msgChildren struct {
	Msg    string
	Children []msgChild
}

func (t *Thing) getChildren(p *Packet) {
	resp := msgChildren{
		Msg: "ReplyChildren",
	}
	for _, child := range t.children {
		resp.Children = append(resp.Children, msgChild{child.id,
			child.model, child.name, child.status})
	}
	t.Reply(p.Marshal(&resp))
}

func (t *Thing) GetChild(id string) *Thing {
	// TODO need R/W lock for t.children[] map
	if child, ok := t.children[id]; ok {
		return child
	}
	return nil
}

// Subscribe to message
func (t *Thing) Subscribe(msg string, f func(*Packet)) {
	t.log.Printf("Subscribed to \"%s\"", msg)
	t.bus.subscribe(msg, f)
}

// Unsubscribe to message
func (t *Thing) Unsubscribe(msg string, f func(*Packet)) {
	if err := t.bus.unsubscribe(msg, f); err != nil {
		t.log.Println(err)
		return
	}
	t.log.Printf("Unsubscribed to \"%s\"", msg)
}

func (t *Thing) receive(p *Packet) {
	t.log.Printf("Received [%s]: %.80s", p.src.Name(), p.String())
	if err := t.bus.receive(p); err != nil {
		t.log.Println(err)
	}
}

func (t *Thing) NewPacket(msg interface {}) *Packet {
	return newPacket(nil, msg)
}

func (t *Thing) Reply(p *Packet) {
	if err := t.bus.reply(p); err != nil {
		t.log.Println(err)
		return
	}
	t.log.Println("Reply", p.String())
}

func (t *Thing) Broadcast(p *Packet) {
	if err := t.bus.broadcast(p); err != nil {
		t.log.Println(err)
		return
	}
	t.log.Printf("Broadcast: %.80s", p.String())
}

// Start the Thing
func (t *Thing) Start() {
	if t.demoMode {
		t.log.Println("Demo mode ENABLED")
	}

	t.httpInit()

	if t.Init != nil {
		t.log.Println("Init...")
		if err := t.Init(); err != nil {
			t.log.Fatalln("Init failed:", err)
		}
	}

	t.tunnelCreate()

	t.httpStart()

	if t.Run != nil {
		t.log.Println("Run...")
		t.Run()
	}

	t.httpStop()

	t.log.Fatalln("Run() didn't run forever")
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

// DefaultId returns a default ID based on system MAC address
func defaultId() string {

	// Use the MAC address of the first non-lo interface

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
	Msg    string
	Id     string
	Model  string
	Name   string
	Status string
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
	t.Broadcast(t.NewPacket(&spam))

	if t.childStatus != nil {
		t.childStatus(child)
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
		t.log.Println("Sorry, you can't be your own Mother")
		goto disconnect
	}

	child = t.GetChild(resp.Id)

	if child == nil {
		child = t.stork(resp.Id, resp.Model, resp.Name)
		if child == nil {
			t.log.Println("Model", resp.Model, "unknown")
			goto disconnect
		}
		t.children[resp.Id] = child
	} else {
		if child.model != resp.Model {
			t.log.Println("Model mismatch")
			goto disconnect
		}
		if child.name != resp.Name {
			t.log.Println("Name mismatch")
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

func (t *Thing) ListenForChildren(max uint, match string, status func(*Thing)) error {
	// TODO thing filter
	t.log.Println("Listening for Children...")
	t.childStatus = status
	t.Subscribe("GetChildren", t.getChildren)
	t.muxPrivate.HandleFunc("/port/{id}", t.getPort)

	return t.portScan(max, match)
}
