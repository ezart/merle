package merle

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type IModel interface {
	Init(*Device, bool) error
	Run()
	Receive(*Packet)
	HomePage(http.ResponseWriter, *http.Request)
}

type Device struct {
	m           IModel
	status      string
	id          string
	model       string
	name        string
	startupTime time.Time
	sync.Mutex
	conns map[*websocket.Conn]bool
}

func NewDevice(m IModel, id, model, name, status string, startupTime time.Time) *Device {
	if id == "" {
		id = DefaultId()
	}

	return &Device{
		m:           m,
		status:      status,
		id:          id,
		model:       model,
		name:        name,
		startupTime: startupTime,
		conns:       make(map[*websocket.Conn]bool),
	}
}

func (d *Device) Id() string {
	return d.id
}

func (d *Device) Model() string {
	return d.model
}

func (d *Device) Name() string {
	return d.name
}

type homeParams struct {
	Scheme string
	Host   string
	Status string
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
		Status: d.status,
		Id:     d.id,
		Model:  d.model,
		Name:   d.name,
	}
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

func (d *Device) Send(p *Packet) {
	log.Printf("Device SendPacket: %s", p.Msg)

	d.Lock()
	defer d.Unlock()

	err := p.writeMessage()
	if err != nil {
		log.Println("Device SendPacket error:", err)
	}
}

func (d *Device) Broadcast(msg []byte) {
	var p = &Packet{
		Msg: msg,
	}

	d.Lock()
	defer d.Unlock()

	if len(d.conns) == 0 {
		log.Printf("Would broadcast: %s", msg)
		return
	}

	log.Printf("Device broadcast: %s", msg)

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then sent on each connection

	for c := range d.conns {
		p.conn = c
		p.writeMessage()
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
			Status:      d.status,
			Id:          d.id,
			Model:       d.model,
			Name:        d.name,
			StartupTime: d.startupTime,
		}
		p.Msg, _ = json.Marshal(&resp)
		d.Send(p)
	default:
		d.m.Receive(p)
	}
}

func (d *Device) receive(p *Packet) {
	var msg MsgType

	log.Printf("Device receivePacket: %s", p.Msg)

	json.Unmarshal(p.Msg, &msg)

	switch msg.Type {
	case MsgTypeCmd:
		d.receiveCmd(p)
	default:
		d.m.Receive(p)
	}
}

func (d *Device) Run(authUser, hubHost, hubUser, hubKey string) {
	err := d.m.Init(d, false)
	if err != nil {
		return
	}

	go d.tunnelCreate(hubHost, hubUser, hubKey)
	go d.http(authUser)

	d.m.Run()
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
