package merle

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type port struct {
	sync.Mutex
	port              uint
	tunnelTrying      bool
	tunnelTryingUntil time.Time
	tunnelConnected   bool
	ws                *websocket.Conn
}

type ports struct {
	max     uint
	begin   uint
	end     uint
	num     uint
	next    uint
	match   string
	ticker  *time.Ticker
	done    chan bool
	ports   []port
	portMap map[string]*port
}

func newPorts(max uint, match string) *ports {
	return &ports{
		max:     max,
		match:   match,
		done:    make(chan bool),
		portMap: make(map[string]*port),
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
			log.Printf("Port[%d] still tunnelTrying", port.port)
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

func (p *ports) init() error {
	var err error

	if p.max == 0 {
		return fmt.Errorf("Max ports equal zero; nothing to scan")
	}

	p.begin, p.end, err = getPortRange()
	if err != nil {
		return err
	}

	p.num = p.end - p.begin + 1
	if p.num > p.max {
		p.num = p.max
		p.end = p.begin + p.num - 1
	}

	p.next = 0

	p.ports = make([]port, p.num)

	for i := uint(0); i < p.num; i++ {
		p.ports[i].port = p.begin + i
	}

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
				log.Printf("Tunnel connected on Port[%d]", port.port)
				port.tunnelConnected = true
				port.tunnelTrying = false
				//go port.attach(match, cb)
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
					log.Println("Scanning ports error:", err)
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
