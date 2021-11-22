package merle

import (
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"net"
	"sync"
	"time"
)

type IDevice interface {
	Init(bool) error
	Run(authUser, hubHost, hubUser, hubKey string)
	ReceivePacket(* Packet)
}

type DeviceGenerator func(id, model, name string, startupTime time.Time) IDevice

type Device struct {
	Id		string
	Model		string
	Name		string
	StartupTime	time.Time

	Input		chan *Packet
	Home		func(http.ResponseWriter, *http.Request)

	sync.Mutex
	conns		map[*websocket.Conn]bool
	//port		*Port
}

func DefaultId() string {

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

type homeParams struct {
	Scheme string
	Host   string
	Id     string
	Model  string
	Name   string
}

func (d *Device) HomeParams(r *http.Request) *homeParams {
	scheme := "wss:\\"
	if r.TLS == nil {
		scheme = "ws:\\"
	}

	return &homeParams{
		Scheme: scheme,
		Host:   r.Host,
		Id:     d.Id,
		Model:  d.Model,
		Name:   d.Name,
	}
}

func (d *Device) homePage(w http.ResponseWriter, r *http.Request) {
	if d.Home == nil {
		fmt.Fprintf(w, "Missing home page handler")
		return
	}

	d.Home(w, r)
}

func (d *Device) connAdd(c *websocket.Conn) {
	d.Lock()
	defer d.Unlock()
	d.conns[c] = true
}

func (d *Device) connDelete(c *websocket.Conn) {
	d.Lock()
	defer d.Unlock()
	delete(d.conns, c)
}

func (d *Device) SendPacket(p *Packet) {
	log.Printf("Device SendPacket: %s", p.Msg)
	err := p.writeMessage()
	if err != nil {
		log.Println("Device SendPacket error:", err)
	}
}

func (d *Device) Broadcast(msg []byte) {
	log.Printf("Device broadcast: %s", msg)

	d.Lock()
	defer d.Unlock()

	if len(d.conns) == 0 {
		log.Printf("Would broadcast: %s", msg)
		return
	}

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then sent on each connection

	for c := range d.conns {
		p := Packet{
			conn: c,
			Msg: msg,
		}
		d.SendPacket(&p)
	}
}

func (d *Device) receiveCmd(p *Packet) {
	var msg MsgCmd

	json.Unmarshal(p.Msg, &msg)

	switch msg.Cmd {

	case CmdIdentify:
		var resp = MsgIdentifyResp{
			Type:        MsgTypeCmdResp,
			Cmd:         msg.Cmd,
			Id:          d.Id,
			Model:       d.Model,
			Name:        d.Name,
			StartupTime: d.StartupTime,
		}
		p.Msg, _ = json.Marshal(&resp)
		d.SendPacket(p)
	default:
		d.Input <- p
	}
}

func (d *Device) receivePacket(p *Packet) {
	var msg MsgType

	log.Printf("Device receivePacket: %s", p.Msg)

	json.Unmarshal(p.Msg, &msg)

	switch msg.Type {
	case MsgTypeCmd:
		d.receiveCmd(p)
	default:
		d.Input <- p
	}
}

func (d *Device) Run(authUser, hubHost, hubUser, hubKey string) {
	d.Input = make(chan *Packet)
	d.conns = make(map[*websocket.Conn]bool)

	go d.tunnelCreate(hubHost, hubUser, hubKey)
	go d.http(authUser)
}
