package merle

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/gorilla/websocket"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type portAttachCb func(*port, *msgIdentity) error

type port struct {
	thing *Thing
	sync.Mutex
	port              uint
	tunnelTrying      bool
	tunnelTryingUntil time.Time
	tunnelConnected   bool
	ws                *websocket.Conn
	done     chan bool
	attachCb portAttachCb
}

func newPort(thing *Thing, p uint, attachCb portAttachCb) *port {
	return &port{
		thing:    thing,
		port:     p,
		done:     make(chan bool),
		attachCb: attachCb,
	}
}

type ports struct {
	thing    *Thing
	begin    uint
	end      uint
	num      uint
	next     uint
	ticker   *time.Ticker
	done     chan bool
	ports    []port
	portMap  map[string]*port
	attachCb portAttachCb
}

func newPorts(thing *Thing, begin, end uint, attachCb portAttachCb) *ports {
	return &ports{
		thing:    thing,
		begin:    begin,
		end:      end,
		done:     make(chan bool),
		portMap:  make(map[string]*port),
		attachCb: attachCb,
	}
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
	msg := struct{ Msg string }{Msg: "_GetIdentity"}
	p.thing.log.Printf("Sending: %.80s", msg)
	return p.ws.WriteJSON(&msg)
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

	p.thing.log.Printf("Received: %.80s", identity)
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
		return nil, errors.Wrap(err, "Websocket open error")
	}

	err = p.wsIdentity()
	if err != nil {
		return nil, errors.Wrap(err, "Send request for Identity failed")
	}

	resp, err = p.wsReplyIdentity()
	if err != nil {
		return nil, errors.Wrap(err, "Didn't reply with Identity in a reasonable time")
	}

	return resp, nil
}

func (p *port) disconnect() {
	p.wsClose()
	p.Lock()
	p.tunnelConnected = false
	p.Unlock()
}

func (p *port) attach() {
	defer p.disconnect()
	resp, err := p.connect()
	if err != nil {
		p.thing.log.Printf("Port[%d] connect failure: %s", p.port, err)
		return
	}

	err = p.attachCb(p, resp)
	if err != nil {
		p.thing.log.Printf("Port[%d] attach failed: %s", p.port, err)
	}
}

// listeningPorts are ports in the range [begin, end] with an active listener.
// An active listener is a Merle tunnel end-point port.
func listeningPorts(begin, end uint) (map[uint]bool, error) {
	listeners := make(map[uint]bool)

	// ss -Hntl4p src 127.0.0.1 sport ge 8081 sport le 9080

	args := []string{
		"-Hntl4",
		"src", "127.0.0.1",
		"sport", "ge", strconv.FormatUint(uint64(begin), 10),
		"sport", "le", strconv.FormatUint(uint64(end), 10),
	}

	cmd := exec.Command("ss", args...)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return listeners, err
	}

	ss := string(stdoutStderr)
	ss = strings.TrimSuffix(ss, "\n")

	for _, ssLine := range strings.Split(ss, "\n") {
		if len(ssLine) > 0 {
			portStr := strings.Split(strings.Split(ssLine,
				":")[1], " ")[0]
			port, _ := strconv.Atoi(portStr)
			listeners[uint(port)] = true
		}
	}

	return listeners, nil
}

func (p *port) scan() error {
	listeners, err := listeningPorts(p.port, p.port)
	if err != nil {
		return err
	}

	p.Lock()
	defer p.Unlock()

	if listeners[p.port] {
		if p.tunnelConnected {
			// no change
		} else {
			p.thing.log.Printf("Tunnel connected on Port[%d]", p.port)
			p.tunnelConnected = true
			go p.attach()
		}
	} else {
		if p.tunnelConnected {
			p.thing.log.Printf("Closing tunnel on Port[%d]", p.port)
			p.tunnelConnected = false
		} else {
			// no change
		}
	}

	return nil
}

func (p *port) run() error {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			if err := p.scan(); err != nil {
				p.thing.log.Println("Scanning port error:", err)
				return err
			}
		}
	}

	return nil
}

func (p *ports) nextPort() (port *port) {

	for i := uint(0); i < p.num; i++ {
		port = &p.ports[p.next]
		p.next++
		if p.next >= p.num {
			p.next = 0
		}
		port.Lock()
		if port.tunnelConnected {
			port.Unlock()
			continue
		}
		if port.tunnelTrying && port.tunnelTryingUntil.After(time.Now()) {
			port.Unlock()
			p.thing.log.Printf("Port[%d] still tunnelTrying", port.port)
			continue
		}
		port.tunnelTrying = true
		port.tunnelTryingUntil = time.Now().Add(2 * time.Second)
		port.Unlock()
		return
	}

	// No more ports
	return nil
}

func (p *ports) getPort(id string) int {
	var port *port
	var ok bool

	if port, ok = p.portMap[id]; ok {
		port.Lock()
		if port.tunnelConnected {
			port.Unlock()
			return -2 // Port busy; try later
		}
		port.Unlock()
	} else {
		port = p.nextPort()
		if port == nil {
			return -1 // No more ports; try later
		}
		p.portMap[id] = port
	}

	return int(port.port)
}

func (p *ports) init() error {
	if p.begin == 0 {
		return fmt.Errorf("Begin port is zero")
	}
	if p.begin > p.end {
		return fmt.Errorf("Begin port %d greater than End port %d", p.begin, p.end)
	}

	p.num = p.end - p.begin + 1

	p.next = 0

	p.ports = make([]port, p.num)

	for i := uint(0); i < p.num; i++ {
		p.ports[i].port = p.begin + i
		p.ports[i].thing = p.thing
	}

	p.thing.log.Printf("Bridge ports[%d-%d]", p.begin, p.end)

	return nil
}

func (p *ports) scan() error {

	listeners, err := listeningPorts(p.begin, p.end)
	if err != nil {
		return err
	}

	for i := uint(0); i < p.num; i++ {
		port := &p.ports[i]
		port.Lock()
		if listeners[port.port] {
			if port.tunnelConnected {
				// no change
			} else {
				p.thing.log.Printf("Tunnel connected on Port[%d]", port.port)
				port.tunnelConnected = true
				port.tunnelTrying = false
				go port.attach()
			}
		} else {
			if port.tunnelConnected {
				p.thing.log.Printf("Closing tunnel on Port[%d]", port.port)
				port.tunnelConnected = false
			} else {
				// no change
			}
		}
		port.Unlock()
	}

	return nil
}

func (p *ports) start() error {
	if err := p.init(); err != nil {
		return err
	}

	p.ticker = time.NewTicker(time.Second)

	go func() {
		for {
			select {
			case <-p.done:
				return
			case <-p.ticker.C:
				if err := p.scan(); err != nil {
					p.thing.log.Println("Scanning ports error:", err)
					return
				}
			}
		}
	}()

	return nil
}

func (p *ports) stop() {
	p.ticker.Stop()
	p.done <- true
}
