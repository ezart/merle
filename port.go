// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a GPLv3 license that can be found in
// the LICENSE file.

package merle

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var portBegin int
var portEnd int
var portNext int

var numPorts int

type port struct {
	sync.Mutex
	port              int
	tunnelTrying      bool
	tunnelTryingUntil time.Time
	tunnelConnected   bool
	ws                *websocket.Conn
}

var ports []port

func (p *port) writeJSON(v interface{}) error {
	log.Printf("Port writeJSON: %v", v)
	return p.ws.WriteJSON(v)
}

func (p *port) writeMessage(msg []byte) error {
	log.Printf("Port writeMessage: %.80s", msg)
	return p.ws.WriteMessage(websocket.TextMessage, msg)
}

func (p *port) readMessage() (msg []byte, err error) {
	_, msg, err = p.ws.ReadMessage()
	if err == nil {
		log.Printf("Port readMessage: %.80s", msg)
	}
	return msg, err
}

func (p *port) wsOpen() error {
	var err error

	u := url.URL{Scheme: "ws",
		Host: "localhost:" + strconv.Itoa(p.port),
		Path: "/ws"}

	p.ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)

	return err
}

func (p *port) wsIdentify() error {
	msg := MsgCmd{MsgTypeCmd, CmdIdentify}
	return p.writeJSON(msg)
}

func (p *port) wsIdentifyResp() (r *MsgIdentifyResp, err error) {
	var resp MsgIdentifyResp

	// Wait for response no longer than a second
	p.ws.SetReadDeadline(time.Now().Add(time.Second))

	err = p.ws.ReadJSON(&resp)
	if err != nil {
		return nil, err
	}

	log.Printf("Port wsIdentifyResp: %v", resp)

	// Clear deadline
	p.ws.SetReadDeadline(time.Time{})

	return &resp, nil
}

func (p *port) wsClose() {
	if p.ws == nil {
		return
	}
	p.ws.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(
			websocket.CloseNormalClosure, ""))
	p.ws.Close()
	p.ws = nil
}

func (p *port) connect() (resp *MsgIdentifyResp, err error) {
	err = p.wsOpen()
	if err != nil {
		log.Printf("Port[%d] run wsOpen error: %s", p.port, err)
		return nil, err
	}

	err = p.wsIdentify()
	if err != nil {
		log.Printf("Port[%d] run wsIdentify error: %s", p.port, err)
		return nil, err
	}

	resp, err = p.wsIdentifyResp()
	if err != nil {
		log.Printf("Port[%d] run wsIdentifyResp error: %s", p.port, err)
		return nil, err
	}

	return resp, nil
}

func (p *port) disconnect() {
	p.wsClose()
	p.Lock()
	p.tunnelConnected = false
	p.Unlock()
}

func (p *port) run(d *Device) {
	var pkt = &Packet{
		conn: p.ws,
	}
	var err error

	d.Lock()
	if d.port != nil {
		d.Unlock()
		log.Printf("Port[%d] already running", p.port)
		return
	}
	d.port = p
	d.Unlock()

	var msg = MsgCmd{
		Type: MsgTypeCmd,
		Cmd:  CmdStart,
	}
	pkt.Msg, _ = json.Marshal(&msg)
	d.m.Receive(pkt)

	for {
		pkt.Msg, err = p.readMessage()
		if err != nil {
			break
		}
		d.m.Receive(pkt)
	}

	d.Lock()
	d.port = nil
	d.Unlock()
}

func (h *Hub) _scanPorts() {
	// ss -Hntl4p src 127.0.0.1 sport ge 8081 sport le 9080

	cmd := exec.Command("ss", "-Hntl4", "src", "127.0.0.1",
		"sport", "ge", strconv.Itoa(portBegin),
		"sport", "le", strconv.Itoa(portEnd))
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Print(err)
		return
	}

	ss := string(stdoutStderr)
	ss = strings.TrimSuffix(ss, "\n")

	listeners := make(map[int]bool)

	for _, ssLine := range strings.Split(ss, "\n") {
		if len(ssLine) > 0 {
			portStr := strings.Split(strings.Split(ssLine,
				":")[1], " ")[0]
			port, _ := strconv.Atoi(portStr)
			listeners[port] = true
		}
	}

	for i := 0; i < numPorts; i++ {
		port := &ports[i]
		port.Lock()
		if listeners[port.port] {
			if port.tunnelConnected {
				// no change
			} else {
				log.Printf("Tunnel connected on port %d", port.port)
				port.tunnelConnected = true
				port.tunnelTrying = false
				go h.portRun(port)
			}
		} else {
			if port.tunnelConnected {
				log.Printf("Closing tunnel on port %d", port.port)
				port.tunnelConnected = false
			} else {
				// no change
			}
		}
		port.Unlock()
	}
}

func (h *Hub) scanPorts() {

	// every second

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			h._scanPorts()
		}
	}
}

func getPortRange() (begin int, end int, err error) {

	c, err := ioutil.ReadFile("/proc/sys/net/ipv4/ip_local_reserved_ports")
	if err != nil {
		return 0, 0, err
	}

	// strip whitespace
	reservedPorts := strings.Fields(string(c))[0]

	// TODO better parsing of reserved ports is needed.  This parser
	// TODO assumes reserved_ports is a single range: [begin-end]

	begin, err = strconv.Atoi(strings.Split(reservedPorts, "-")[0])
	if err != nil {
		return 0, 0, err
	}

	end, err = strconv.Atoi(strings.Split(reservedPorts, "-")[1])
	if err != nil {
		return 0, 0, err
	}

	log.Printf("Port range [%d-%d]", begin, end)

	return begin, end, nil
}

func (h *Hub) portScan() {
	var err error

	portBegin, portEnd, err = getPortRange()
	if err != nil {
		log.Fatal(err)
	}

	numPorts = portEnd - portBegin + 1
	portNext = 0

	ports = make([]port, numPorts)

	for i := 0; i < numPorts; i++ {
		ports[i].port = portBegin + i
	}

	h.scanPorts()
}

func nextPort() (p *port) {

	for i := 0; i < numPorts; i++ {
		p = &ports[portNext]
		portNext++
		if portNext >= numPorts {
			portNext = 0
		}
		p.Lock()
		if p.tunnelConnected {
			p.Unlock()
			continue
		}
		if p.tunnelTrying && p.tunnelTryingUntil.After(time.Now()) {
			p.Unlock()
			log.Printf("NextPort still tunnelTrying on port %d", p.port)
			continue
		}
		p.tunnelTrying = true
		p.tunnelTryingUntil = time.Now().Add(2 * time.Second)
		p.Unlock()
		return
	}

	// No more ports
	return nil
}

var portMap = make(map[string]*port)

func portFromId(id string) int {
	var p *port
	var ok bool

	if p, ok = portMap[id]; ok {
		p.Lock()
		if p.tunnelConnected {
			p.Unlock()
			log.Printf("portFromId ID %s port %d busy", id, p.port)
			return -2 // Port busy; try later
		}
		p.Unlock()
	} else {
		p = nextPort()
		if p == nil {
			log.Printf("portFromId ID %s no more ports", id)
			return -1 // No more ports; try later
		}
		portMap[id] = p
	}

	log.Printf("portFromId ID %s port %d", id, p.port)
	return p.port
}
