package merle

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type bridgeConfig struct {
	Bridge struct {
		Max   uint   `yaml:"Max"`
		Match string `yaml:"Match"`
	} `yaml:"Bridge"`
}

type bridger interface {
	BridgeSubscribe() Subscribers
}

type children map[string]*Thing

type bridge struct {
	bridger  bridger
	children children
	bus      *bus
	ports    *ports
}

func newBridge(bridger bridger, config Configurator) (*bridge, error) {
	var cfg bridgeConfig

	if err := must(config.Parse(&cfg)); err != nil {
		return nil, err
	}

	return &bridge{
		bridger:  bridger,
		children: make(children),
		bus:      newBus(10, bridger.BridgeSubscribe()),
		ports:    newPorts(cfg.Bridge.Max, cfg.Bridge.Match),
	}, nil
}

func (b *bridge) getPort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	port := b.ports.getPort(id)

	switch port {
	case -1:
		fmt.Fprintf(w, "no ports available")
	case -2:
		fmt.Fprintf(w, "port busy")
	default:
		fmt.Fprintf(w, "%d", port)
	}
}

func (b *bridge) Start() {
	must(b.ports.Start())
}

func (b *bridge) Stop() {
	b.ports.Stop()
	b.bus.close()
}
