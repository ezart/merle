// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
	"log"
)

var portBegin uint
var portEnd uint
var portNext uint

var numPorts uint

type port struct {
	sync.Mutex
	port              uint
	tunnelTrying      bool
	tunnelTryingUntil time.Time
	tunnelConnected   bool
	ws                *websocket.Conn
}

var ports []port

func (p *port) writeJSON(v interface{}) error {
	return p.ws.WriteJSON(v)
}

func (p *port) writeMessage(msg []byte) error {
	return p.ws.WriteMessage(websocket.TextMessage, msg)
}

func (p *port) readMessage() (msg []byte, err error) {
	_, msg, err = p.ws.ReadMessage()
	return msg, err
}

func (p *port) wsOpen() error {
	var err error

	u := url.URL{Scheme: "ws",
		Host: "localhost:" + strconv.FormatUint(uint64(p.port), 10),
		Path: "/ws"}

	p.ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)

	return err
}

func (p *port) wsIdentity() error {
	msg := struct{ Msg string }{Msg: "GetIdentity"}
	return p.writeJSON(&msg)
}

func (p *port) wsReplyIdentity() (resp *msgIdentity, err error) {
	var identity msgIdentity

	// Wait for response no longer than a second
	p.ws.SetReadDeadline(time.Now().Add(time.Second))

	err = p.ws.ReadJSON(&identity)
	if err != nil {
		return nil, err
	}

	// Clear deadline
	p.ws.SetReadDeadline(time.Time{})

	return &identity, nil
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

func (p *port) connect() (resp *msgIdentity, err error) {
	err = p.wsOpen()
	if err != nil {
		return nil, fmt.Errorf("Websocket open error: %s", err)
	}

	err = p.wsIdentity()
	if err != nil {
		return nil, fmt.Errorf("Send request for Identity failed: %s", err)
	}

	resp, err = p.wsReplyIdentity()
	if err != nil {
		return nil,
			fmt.Errorf("Didn't reply with Identity in a reasonable time: %s",
			err)
	}

	return resp, nil
}

func (p *port) disconnect() {
	p.wsClose()
	p.Lock()
	p.tunnelConnected = false
	p.Unlock()
}

type portAttach func(*port, *msgIdentity) error

func (p *port) attach(match string, cb portAttach) {
	resp, err := p.connect()
	if err != nil {
		log.Println("Port connect failure:", err)
		goto disconnect
	}

	// TODO disconnect if resp doesn't match filter

	err = cb(p, resp)
	if err != nil {
		log.Println("Port attach failed:", err)
	}

disconnect:
	p.disconnect()
}

func getPortRange() (begin uint, end uint, err error) {

	// Merle uses ip_local_reserved_ports for incoming Thing
	// connections.
	//
	// Set a range using:
	//
	//   sudo sysctl -w net.ipv4.ip_local_reserved_ports="8000-8040"
	//
	// Or, to persist setting on next boot, add to /etc/sysctl.conf:
	//
	//   net.ipv4.ip_local_reserved_ports = 8000-8040
	//
	// And then run sudo sysctl -p
	//
	// Notes:
	//
	//    1) ip_local_reserved_ports range needs to be included in
	//       ip_local_port_range
	//
	//    2) Be careful that Thing.portPrivate is outside
	//       ip_local_reserved_ports range
	//
	//    3) The number of ports defined by ip_local_reserved_ports
	//       range is the max number of incoming connections.  In
	//       the example above, max = (8040 - 8000) + 1

	bytes, err := ioutil.ReadFile("/proc/sys/net/ipv4/ip_local_reserved_ports")
	if err != nil {
		return 0, 0, err
	}

	// strip whitespace
	fields := strings.Fields(string(bytes))
	if len(fields) == 0 {
		return 0, 0, fmt.Errorf("Missing /proc/sys/net/ipv4/ip_local_reserved_ports?")
	}

	// TODO better parsing of reserved ports is needed.  This parser
	// TODO assumes reserved_ports is a single range: [begin-end]

	reservedPorts := fields[0]

	b, err := strconv.Atoi(strings.Split(reservedPorts, "-")[0])
	if err != nil {
		return 0, 0, err
	}
	begin = uint(b)

	e, err := strconv.Atoi(strings.Split(reservedPorts, "-")[1])
	if err != nil {
		return 0, 0, err
	}
	end = uint(e)

	return begin, end, nil
}

func initPorts(max uint) error {
	var err error

	if max == 0 {
		return fmt.Errorf("Max ports equal zero; nothing to scan")
	}

	portBegin, portEnd, err = getPortRange()
	if err != nil {
		return err
	}

	numPorts = portEnd - portBegin + 1
	if numPorts > max {
		numPorts = max
		portEnd = portBegin + numPorts - 1
	}

	portNext = 0

	ports = make([]port, numPorts)

	for i := uint(0); i < numPorts; i++ {
		ports[i].port = portBegin + i
	}

	return nil
}

func _portScan(match string, cb portAttach) error {

	// ss -Hntl4p src 127.0.0.1 sport ge 8081 sport le 9080

	cmd := exec.Command("ss", "-Hntl4", "src", "127.0.0.1",
		"sport", "ge", strconv.FormatUint(uint64(portBegin), 10),
		"sport", "le", strconv.FormatUint(uint64(portEnd), 10))
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Scanning ports error:", err)
		return err
	}

	ss := string(stdoutStderr)
	ss = strings.TrimSuffix(ss, "\n")

	listeners := make(map[uint]bool)

	for _, ssLine := range strings.Split(ss, "\n") {
		if len(ssLine) > 0 {
			portStr := strings.Split(strings.Split(ssLine,
				":")[1], " ")[0]
			port, _ := strconv.Atoi(portStr)
			listeners[uint(port)] = true
		}
	}

	for i := uint(0); i < numPorts; i++ {
		port := &ports[i]
		port.Lock()
		if listeners[port.port] {
			if port.tunnelConnected {
				// no change
			} else {
				log.Printf("Tunnel connected on Port[%d]", port.port)
				port.tunnelConnected = true
				port.tunnelTrying = false
				go port.attach(match, cb)
			}
		} else {
			if port.tunnelConnected {
				log.Println("Closing tunnel on Port[%d]", port.port)
				port.tunnelConnected = false
			} else {
				// no change
			}
		}
		port.Unlock()
	}

	return nil
}

func portScan(max uint, match string, cb portAttach) error {

	if err := initPorts(max); err != nil {
		return err
	}

	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := _portScan(match, cb); err != nil {
					return
				}
			}
		}
	}()

	return nil
}

func nextPort() (p *port) {

	for i := uint(0); i < numPorts; i++ {
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
			log.Printf("Port[%d] still tunnelTrying", p.port)
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
			log.Printf("Port[%d] busy; already used by Id %s", p.port, id)
			return -2 // Port busy; try later
		}
		p.Unlock()
	} else {
		p = nextPort()
		if p == nil {
			log.Println("No more ports Id", id)
			return -1 // No more ports; try later
		}
		portMap[id] = p
	}

	log.Println("Returning port", p.port, "for Id", id)
	return int(p.port)
}
