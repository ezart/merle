package merle

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"net/http"
	"path"
	"time"
)

type ThingAssets struct {
	Dir string
	Template string
}

// All things implement this interface
type Thinger interface {
	// List of subscribers on thing bus.  On packet receipt, the
	// subscribers are process in-order, and the first matching subscriber
	// stops the processing.
	Subscribers() Subscribers
	Assets() *ThingAssets
}

type Thingers map[string]func() Thinger

// Thing's backing structure
type Thing struct {
	thinger     Thinger
	cfg         *ThingConfig
	assets      *ThingAssets
	status      string
	id          string
	model       string
	name        string
	startupTime time.Time
	bus         *bus
	tunnel      *tunnel
	private     *webPrivate
	public      *webPublic
	templ       *template.Template
	templErr    error
	isBridge    bool
	bridge      *bridge
	isPrime     bool
	primePort   *port
	primeId     string
	log         *log.Logger
}

func NewThing(thinger Thinger, cfg *ThingConfig) *Thing {
	id := cfg.Thing.Id
	isPrime := cfg.Thing.Prime

	if !isPrime {
		if id == "" {
			id = defaultId()
			log.Println("Defaulting ID to", id)
		}
	}

	prefix := "[" + id + "] "

	t := &Thing{
		thinger:     thinger,
		cfg:         cfg,
		assets:      thinger.Assets(),
		status:      "online",
		id:          id,
		model:       cfg.Thing.Model,
		name:        cfg.Thing.Name,
		startupTime: time.Now(),
		isPrime:     isPrime,
		log:         log.New(os.Stderr, prefix, 0),
	}

	t.bus = newBus(t, 10, thinger.Subscribers())

	t.tunnel = newTunnel(t.id, cfg.Mother.Host, cfg.Mother.User,
		cfg.Mother.Key, cfg.Thing.PortPrivate, cfg.Mother.PortPrivate)

	t.private = newWebPrivate(t, cfg.Thing.PortPrivate)
	t.public = newWebPublic(t, cfg.Thing.PortPublic, cfg.Thing.PortPublicTLS,
		cfg.Thing.User)
	t.setAssetsDir(t)

	templ := path.Join(t.assets.Dir, t.assets.Template)
	t.templ, t.templErr = template.ParseFiles(templ)

	_, t.isBridge = t.thinger.(Bridger)
	if t.isBridge {
		t.bridge = newBridge(t)
	}

	if t.isPrime {
		t.private.handleFunc("/port/{id}", t.getPrimePort)
		t.primePort = newPort(t, cfg.Thing.PortPrime, t.primeAttach)
	}

	t.bus.subscribe("_GetIdentity", t.getIdentity)

	return t
}

type msgIdentity struct {
	Msg         string
	Status      string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}

func (t *Thing) getIdentity(p *Packet) {
	resp := msgIdentity{
		Msg:         "_ReplyIdentity",
		Status:      t.status,
		Id:          t.id,
		Model:       t.model,
		Name:        t.name,
		StartupTime: t.startupTime,
	}
	p.Marshal(&resp).Reply()
}

func (t *Thing) getChild(id string) *Thing {
	if !t.isBridge {
		return nil
	}
	return t.bridge.getChild(id)
}

func (t *Thing) run() error {
	t.private.start()
	t.public.start()
	t.tunnel.start()

	if t.isBridge {
		t.bridge.start()
	}

	msg := struct{ Msg string }{Msg: "_CmdRun"}
	t.bus.receive(newPacket(t.bus, nil, &msg))

	if t.isBridge {
		t.bridge.stop()
	}

	t.tunnel.stop()
	t.public.stop()
	t.private.stop()

	t.bus.close()

	return fmt.Errorf("_CmdRun didn't run forever")
}

func (t *Thing) Run() error {
	switch {
	case t.isPrime:
		return t.runPrime()
	default:
		return t.run()
	}
}

func (t *Thing) runOnPort(p *port) error {
	var name = fmt.Sprintf("port:%d", p.port)
	var sock = newWebSocket(name, p.ws)
	var pkt = newPacket(t.bus, sock, nil)
	var err error

	t.log.Printf("Websocket opened [%s]", name)

	t.bus.plugin(sock)

	msg := struct{ Msg string }{Msg: "_CmdRunPrime"}
	t.bus.receive(pkt.Marshal(&msg))

	for {
		// new pkt for each rcv
		var pkt = newPacket(t.bus, sock, nil)

		pkt.msg, err = p.readMessage()
		if err != nil {
			t.log.Printf("Websocket closed [%s]", name)
			break
		}
		t.bus.receive(pkt)
	}

	t.bus.unplug(sock)

	return err
}

func (t *Thing) setAssetsDir(child *Thing) {
	fs := http.FileServer(http.Dir(child.assets.Dir))
	t.public.mux.PathPrefix("/" + child.id + "/assets/").
		Handler(http.StripPrefix("/" + child.id + "/assets/", fs))
}
