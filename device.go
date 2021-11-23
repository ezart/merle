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
	Init(*Device) error
	Run()
	Receive([]byte)
	HomePage(http.ResponseWriter, *http.Request)
}

type Device struct {
	m           IModel
	id          string
	model       string
	name        string
	startupTime time.Time

	sync.Mutex
	conns map[*websocket.Conn]bool
}

func NewDevice(m IModel, id, model, name string) *Device {
	return &Device{
		m: m,
		id: id,
		model: model,
		name: name,
		startupTime: time.Now(),
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
			Msg:  msg,
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
			Id:          d.id,
			Model:       d.model,
			Name:        d.name,
			StartupTime: d.startupTime,
		}
		p.Msg, _ = json.Marshal(&resp)
		d.SendPacket(p)
	default:
		//d.Input <- p
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
		//d.Input <- p
	}
}

func (d *Device) Run(authUser, hubHost, hubUser, hubKey string) {
	//d.Input = make(chan *Packet)
	d.conns = make(map[*websocket.Conn]bool)

	go d.tunnelCreate(hubHost, hubUser, hubKey)
	go d.http(authUser)
}
