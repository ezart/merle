package merle

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type Thing struct {
	Init func() error
	Run  func()
	Home func(w http.ResponseWriter, r *http.Request)
	childFromId func(string) *Thing

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

	stork       func(string, string, string) (*Thing, error)

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

func (t *Thing) SetStork(f func(string, string, string) (*Thing, error)) {
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

func (t *Thing) InitThing(id, model, name string) (*Thing, error) {
	if t.inited {
		return nil, fmt.Errorf("%sAlready inited!", t.logPrefix())
	}

	if id == "" {
		t.id = defaultId()
	} else {
		t.id = id
	}

	t.model = model
	t.name = name

	t.log = log.New(os.Stderr, t.logPrefix(), 0)

	if id == "" {
		t.log.Println("Thing Id is missing; defaulting to", t.id)
	}
	if model == "" {
		return nil, fmt.Errorf("Thing Model is missing")
	}
	if name == "" {
		return nil, fmt.Errorf("Thing Name is missing")
	}

	t.status = "online"
	t.startupTime = time.Now()

	// TODO pass in connMax from cfg?
	t.bus = NewBus(10)

	t.stork = func(string, string, string) (*Thing, error) {
		return nil, fmt.Errorf("need to set stork")
	}

	t.childFromId = func(string) *Thing {
		return nil
	}

	t.httpInit()

	t.Subscribe("GetIdentity", t.getIdentity)

	t.inited = true

	return t, nil
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

func (t *Thing) plugin(sock ISocket) {
	t.log.Println("Plugged into socket", sock.Name())
	t.bus.plugin(sock)
}

func (t *Thing) unplug(sock ISocket) {
	t.log.Println("Unplugged from socket", sock.Name())
	t.bus.unplug(sock)
}

func (t *Thing) receive(p *Packet) {
	t.log.Printf("Received [%s]: %.80s", p.src.Name(), p.String())
	if err := t.bus.receive(p); err != nil {
		t.log.Println(err)
	}
}

func (t *Thing) NewPacket(msg interface{}) *Packet {
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

	if t.Init != nil {
		t.log.Println("Init...")
		if err := t.Init(); err != nil {
			t.log.Fatalln("Init failed:", err)
		}
	}

	t.tunnelCreate()

	t.httpStart()

	t.log.Println("Run...")
	if t.Run == nil {
		select {}
	} else {
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

func (t *Thing) run(p *port) {
	var name = fmt.Sprintf("port:%d", p.port)
	var sock = newWebSocket(name, p.ws)
	var pkt = newPacket(sock, nil)
	var err error

	t.plugin(sock)

	msg := struct{ Msg string }{Msg: "CmdStart"}
	t.receive(pkt.Marshal(&msg))

	for {
		// new pkt for each rcv
		var pkt = newPacket(sock, nil)

		pkt.msg, err = p.readMessage()
		if err != nil {
			break
		}
		t.receive(pkt)
	}

	t.unplug(sock)
}

