// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in
// the LICENSE file.

package merle

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// IModel is the business logic of a Device, specifying a "device driver" interface.
type IModel interface {

	// Init the device model.  Device model software construct such as
	// timers/tickers can be setup here.  If running in device-mode, device
	// model hardware is initialized here as well.
	Init(d *Device) error

	// Run the device model.  Run should block until device stops.  Run is
	// only called in device-mode.  Hub-mode will not call Run.
	Run()

	// Receive a message in a Packet.
	Receive(p *Packet)

	// HomePage for the device model.
	HomePage(w http.ResponseWriter, r *http.Request)
}

// Device runs an IModel, either in device-mode or hub-mode.  In device-mode,
// the Device runs locally, with direct access to the device hardware.  In
// hub-mode, the Device runs inside a Hub, connecting to the device-mode Device
// over a tunnel using websockets.  The IModel implements both modes.
//
// Device-mode
//
// In device-mode, the Device runs the IModel, starts two web servers, and
// (optionally) creates a tunnel to a Hub.
//
// The first web server listens on port :80 for HTTP and serves up the Device
// homepage on http://localhost/ and a websocket connection to the Device on
// ws://localhost/ws.  Basic Authentication protects access to both http:// and
// ws://.  The only allowed Basic Authentication user is authUser passed to
// Device.Run(...).
//
// The second web server listens on port :8080 for HTTP and serves up websocket
// connection to the Device on ws://localhost:8080/ws.
//
// The optional tunnel is a SSH remote port forwarding tunnel where the Device
// is the ssh client and the Hub is the SSH server.  To create the tunnel,
// first the Device requests, over SSH, a remote port from the Hub.  The Device
// then creates a SSH remote port forwarding tunnel mapping <remote
// port>:localhost:8080.  The Hub can now create a websocket connection back to
// the device-mode Device using ws://localhost:<remote port>/ws.
//
// Hub-mode
//
// In hub-mode, the Device runs the IModel inside a Hub.  The Hub will
// websocket connect back to the device-mode Device, also running the same
// IModel.  See type Hub for more information.
// 
type Device struct {
	sync.Mutex
	m           IModel
	status      string
	id          string
	model       string
	name        string
	startupTime time.Time
	conns       map[*websocket.Conn]bool
	port        *port
	inHub       bool
}

// NewDevice returns a new Device.
// 	m is IModel instance.
// 	inHub is true if Device is running in hub-mode.
// 	id is ID of Device.  id is unique for each Device in a Hub.  If id is "",
// 	a default ID is assigned.
// 	model is the name of the model.
// 	name is the name of the Device.
// 	status is status of the Device, e.g. "online", "offline".
// 	startupTime is the Device's startup time.
func NewDevice(m IModel, inHub bool, id, model, name, status string,
	startupTime time.Time) *Device {
	if id == "" {
		id = DefaultId()
	}
	if model == "" || name == "" || status == "" {
		return nil
	}

	return &Device{
		m:           m,
		status:      status,
		id:          id,
		model:       model,
		name:        name,
		startupTime: startupTime,
		inHub:       inHub,
		conns:       make(map[*websocket.Conn]bool),
	}
}

// Return the Device ID
func (d *Device) Id() string {
	return d.id
}

// Return the Device model
func (d *Device) Model() string {
	return d.model
}

// Return the Device name
func (d *Device) Name() string {
	return d.name
}

// Return true if Device running in Hub
func (d *Device) InHub() bool {
	return d.inHub
}

type homeParams struct {
	Scheme string
	Host   string
	Status string
	Id     string
	Model  string
	Name   string
}

// HomeParams returns useful parameters to be passed to IModel.HomePage's
// html template.
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

// Reply sends Packet back to originator of a Packet received with
// IModel.Receive.
func (d *Device) Reply(p *Packet) {
	log.Printf("Device Reply: %.80s", p.Msg)

	d.Lock()
	defer d.Unlock()

	err := p.writeMessage()
	if err != nil {
		log.Println("Device Reply error:", err)
	}
}

// Sink sends Packet towards device-mode Device.  The Packet is not sunk if
// Device is not in hub-mode.  The Packet is not sunk if it came thru the port
// from the device-mode Device.
func (d *Device) Sink(p *Packet) {
	if !d.inHub {
		return
	}

	src := p.conn

	d.Lock()
	defer func() {
		p.conn = src
		d.Unlock()
	}()

	if d.port == nil {
		log.Printf("Device Sink error: not running on port: %s", p.Msg)
		return
	}

	if src == d.port.ws {
		log.Printf("Device Sink reject: message came in on port: %s", p.Msg)
		return
	}

	log.Printf("Device Sink: %.80s", p.Msg)

	p.conn = d.port.ws

	err := p.writeMessage()
	if err != nil {
		log.Println("Device Sink error:", err)
	}
}

// Broadcast packet to all websocket connections on the Device, except self.
func (d *Device) Broadcast(p *Packet) {
	src := p.conn

	d.Lock()
	defer func() {
		p.conn = src
		d.Unlock()
	}()

	switch len(d.conns) {
	case 0:
		log.Printf("Would broadcast: %.80s", p.Msg)
		return
	case 1:
		if _, ok := d.conns[src]; ok {
			log.Printf("Would broadcast: %.80s", p.Msg)
			return
		}
	}

	log.Printf("Device broadcast: %.80s", p.Msg)

	// TODO Perf optimization: use websocket.NewPreparedMessage
	// TODO to prepare msg once, and then send on each connection

	for c := range d.conns {
		if c == src {
			// skip self
			log.Printf("Skipping broadcast: %.80s", p.Msg)
			continue
		}
		p.conn = c
		log.Printf("Sending broadcast: %.80s", p.Msg)
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
		d.Reply(p)
	default:
		d.m.Receive(p)
	}
}

func (d *Device) receive(p *Packet) {
	var msg MsgType

	log.Printf("Device receivePacket: %.80s", p.Msg)

	json.Unmarshal(p.Msg, &msg)

	switch msg.Type {
	case MsgTypeCmd:
		d.receiveCmd(p)
	default:
		d.m.Receive(p)
	}
}

// Run the Device.  Run should not be called on a Device in Hub.  Run will
// initialize the IModel, create a tunnel, start the http servers, and then run
// the IModel.
// 	authUser is the valid user for Basic Authentication.
// 	hubHost is URL for the Hub host.  If blank, Device will not connect to Hub.
//	hubUser is the Hub SSH user.
// 	hubKey is the Hub SSH key.
func (d *Device) Run(authUser, hubHost, hubUser, hubKey string) error {
	if d.inHub {
		return nil
	}

	err := d.m.Init(d)
	if err != nil {
		return err
	}

	go d.tunnelCreate(hubHost, hubUser, hubKey)
	go d.http(authUser)

	d.m.Run()

	return fmt.Errorf("Device Run() exited unexpectedly")
}

// DefaultId returns a default ID based on the device's MAC address
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
