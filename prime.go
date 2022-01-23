package merle

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func (t *Thing) primeAttach(p *port, msg *msgIdentity) error {
	var sockname = fmt.Sprintf("port:%d", p.port)
	var sock = newWebSocket(sockname, p.ws)
	var err error

	if msg.Model != t.cfg.Thing.Model {
		return fmt.Errorf("Model mis-match: want %s, got %s",
			t.cfg.Thing.Model, msg.Model)
	}

	t.id = msg.Id
	t.model = msg.Model
	t.name = msg.Name
	t.startupTime = msg.StartupTime

	prefix := "[" + t.id + "] "
	t.log = log.New(os.Stderr, prefix, 0)

	t.primeId = t.id

	t.bus.plugin(sock)

	// Send a _CmdRunPrime message on startup to Thing-prime so Thing-prime
	// can get current state from the real Thing
	msgRunPrime := struct{ Msg string }{Msg: "_CmdRunPrime"}
	pkt := newPacket(t.bus, sock, nil)
	t.bus.receive(pkt.Marshal(&msgRunPrime))

	for {
		// new pkt for each rcv
		var pkt = newPacket(t.bus, sock, nil)

		pkt.msg, err = p.readMessage()
		if err != nil {
			break
		}
		t.bus.receive(pkt)
	}

	t.bus.unplug(sock)

	return err
}

func (t *Thing) getPrimePort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	t.primePort.Lock()
	defer t.primePort.Unlock()

	if t.primePort.tunnelConnected {
		fmt.Fprintf(w, "port busy")
		return
	}

	if t.primeId != "" && t.primeId != id {
		fmt.Fprintf(w, "no ports available")
		return
	}

	fmt.Fprintf(w, "%d", t.primePort.port)
}

func (t *Thing) runPrime() error {
	t.private.start()
	t.public.start()
	t.tunnel.start()
	return t.primePort.run()
}
