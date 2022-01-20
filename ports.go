package merle

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type port struct {
	log *log.Logger
	sync.Mutex
	port              uint
	tunnelTrying      bool
	tunnelTryingUntil time.Time
	tunnelConnected   bool
	ws                *websocket.Conn
}

type portAttachCb func(*port, *msgIdentity) error

type ports struct {
	log      *log.Logger
	begin    uint
	end      uint
	num      uint
	next     uint
	match    string
	ticker   *time.Ticker
	done     chan bool
	ports    []port
	portMap  map[string]*port
	attachCb portAttachCb
}

func newPorts(log *log.Logger, begin, end uint, match string, attachCb portAttachCb) *ports {
	return &ports{
		log:      log,
		begin:    begin,
		end:      end,
		match:    match,
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
	p.log.Printf("Sending: %.80s", msg)
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

	p.log.Printf("Received: %.80s", identity)
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

func (p *port) attach(match string, cb portAttachCb) {
	resp, err := p.connect()
	defer p.disconnect()
	if err != nil {
		p.log.Printf("Port[%d] connect failure: %s", p.port, err)
		return
	}

	spec := resp.Id + ":" + resp.Model + ":" + resp.Name
	matched, err := regexp.MatchString(match, spec)
	if err != nil {
		p.log.Printf("Port[%d] error compiling regexp \"%s\": %s",
			p.port, match, err)
		return
	}

	if !matched {
		p.log.Printf("Port[%d] Thing [%s] didn't match filter [%s]; not attaching",
			p.port, spec, match)
		return
	}

	err = cb(p, resp)
	if err != nil {
		p.log.Printf("Port[%d] attach failed: %s", p.port, err)
	}
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
			p.log.Printf("Port[%d] still tunnelTrying", port.port)
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
		p.ports[i].log = p.log
	}

	p.log.Printf("Bridge ports[%d-%d]", p.begin, p.end)

	return nil
}

func (p *ports) scan() error {

	// ss -Hntl4p src 127.0.0.1 sport ge 8081 sport le 9080

	cmd := exec.Command("ss", "-Hntl4", "src", "127.0.0.1",
		"sport", "ge", strconv.FormatUint(uint64(p.begin), 10),
		"sport", "le", strconv.FormatUint(uint64(p.end), 10))
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
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

	for i := uint(0); i < p.num; i++ {
		port := &p.ports[i]
		port.Lock()
		if listeners[port.port] {
			if port.tunnelConnected {
				// no change
			} else {
				p.log.Printf("Tunnel connected on Port[%d]", port.port)
				port.tunnelConnected = true
				port.tunnelTrying = false
				go port.attach(p.match, p.attachCb)
			}
		} else {
			if port.tunnelConnected {
				p.log.Printf("Closing tunnel on Port[%d]", port.port)
				port.tunnelConnected = false
			} else {
				// no change
			}
		}
		port.Unlock()
	}

	return nil
}

func (p *ports) Start() error {
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
					p.log.Println("Scanning ports error:", err)
					return
				}
			}
		}
	}()

	return nil
}

func (p *ports) Stop() {
	p.ticker.Stop()
	p.done <- true
}
