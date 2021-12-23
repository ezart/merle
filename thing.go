package merle

import (
	"log"
	"time"
	"net"
	"net/http"
	"encoding/json"
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

	handlers      map[string]func(*Packet)
}

func (t *Thing) logPrefix() string {
	return "["+t.Id+","+t.Model+","+t.Name+"]"
}

func (t *Thing) Start() {
	if t.Init != nil {
		log.Println(t.logPrefix(), "Init...")
		if err := t.Init(); err != nil {
			log.Fatalln(t.logPrefix(), "Init failed:", err)
		}
	}
	if t.Run != nil {
		log.Println(t.logPrefix(), "Run...")
		t.Run()
	}
	log.Fatalln(t.logPrefix(), "Run() didn't run forever")
}

func (t *Thing) Sink(p *Packet) {
}

func (t *Thing) Broadcast(p *Packet) {
}

func (t *Thing) NewPacket(msg interface {}) *Packet {
	var p Packet
	p.Msg, _ = json.Marshal(&msg)
	return &p
}

func (t *Thing) AddHandler(msgType string, f func(*Packet)) {
	if t.handlers == nil {
		t.handlers = make(map[string]func(*Packet))
	}
	t.handlers[msgType] = f
}

type homeParams_ struct {
	Scheme string
	Host   string
	Status string
	Id     string
	Model  string
	Name   string
}

func (t *Thing) HomeParams(r *http.Request) *homeParams_ {
	scheme := "wss:\\"
	if r.TLS == nil {
		scheme = "ws:\\"
	}

	return &homeParams_{
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
