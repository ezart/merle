package merle

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func (t *Thing) changeStatus(status string) {
	t.status = status

	spam := SpamStatus{
		Msg:    "_SpamStatus",
		Id:     t.id,
		Model:  t.model,
		Name:   t.name,
		Status: t.status,
	}
	newPacket(t.bus, nil, &spam).Broadcast()
}

func (t *Thing) primeAttach(p *port, msg *msgIdentity) error {
	if msg.Model != t.cfg.Thing.Model {
		return fmt.Errorf("Model mis-match: want %s, got %s",
			t.cfg.Thing.Model, msg.Model)
	}

	t.id = msg.Id
	t.model = msg.Model
	t.name = msg.Name
	t.startupTime = msg.StartupTime
	t.primeId = t.id

	prefix := "[" + t.id + "] "
	t.log = log.New(os.Stderr, prefix, 0)

	t.setAssetsDir(t)

	t.changeStatus("online")
	err := t.runOnPort(p)
	t.changeStatus("offline")

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
