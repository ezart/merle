package merle

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// Thing configuration.  All Things share this configuration.
type ThingConfig struct {

	// The section describes a Thing.
	Thing struct {
		// [Optional] Thing's Id.  Ids are unique within an application
		// to differenciate one Thing from another.  Id is optional; if
		// Id is not given, a system-wide unique Id is assigned.
		Id string `yaml:"Id"`
		// Thing's Model.
		Model string `yaml:"Model"`
		// Thing's Name
		Name string `yaml:"Name"`
		// [Optional] system User.  If a User is given, any browser
		// views of the Thing's home page  will prompt for user/passwd.
		// HTTP Basic Authentication is used and the user/passwd given
		// must match the system creditials for the User.  If no User
		// is given, HTTP Basic Authentication is skipped; anyone can
		// view the home page.
		User string `yaml:"User"`
		// [Optional] If PortPublic is given, an HTTP web server is
		// started on port PortPublic.  PortPublic is typically set to
		// 80.  The HTTP web server runs the Thing's home page.
		PortPublic uint `yaml:"PortPublic"`
		// [Optional] If PortPublicTLS is given, an HTTPS web server is
		// started on port PortPublicTLS.  PortPublicTLS is typically
		// set to 443.  The HTTPS web server will self-certify using a
		// certificate from Let's Encrypt.  The public HTTPS server
		// will securely run the Thing's home page.  If PortPublicTLS
		// is given, PortPublic must be given.
		PortPublicTLS uint `yaml:"PortPublicTLS"`
		// [Optional] If PortPrivate is given, a HTTP server is
		// started on port PortPrivate.  This HTTP server does not
		// server up the Thing's home page but rather connects to
		// Thing's Mother using a websocket over HTTP.
		PortPrivate uint `yaml:"PortPrivate"`
		// [Optional] Run as Thing-prime.
		Prime bool `yaml:"Prime"`
		// [Required, if Prime] PortPrime port is used to create a
		// tunnel from Thing to Thing-prime.  The port should be a
		// reserved port in ip_local_reserved_ports.
		PortPrime uint `yaml:"PortPrime"`
		// [Optional] Web assets directory (location of html/js/css files)
		AssetsDir string `yaml:"AssetsDir"`
	} `yaml:"Thing"`

	// [Optional] This section describes a Thing's Mother.  Every Thing has
	// a Mother.  A Mother is also a Thing.  We can build a hierarchy of
	// Things, with a Thing having a Mother, a GrandMother, a Great
	// GrandMother, etc.
	Mother struct {
		// Mother's Host address.  Host, User and Key are used to
		// connect this Thing to it's Mother using a SSH connection.
		// For example: ssh -i <Key> <User>@<Host>.
		Host string `yaml:"Host"`
		// User on Host associated with Key
		User string `yaml:"User"`
		// Key is the file path of the SSH identity key.  See ssh -i
		// option for more information.
		Key string `yaml:"Key"`
		// Port on Host for Mother's private HTTP server
		PortPrivate uint `yaml:"PortPrivate"`
	} `yaml:"Mother"`

	// [Optional] Bridge configuration.  A Thing implementing the Bridger
	// interface will use this config for bridge-specific configuration.
	Bridge struct {
		// Beginning port number.  The bridge will listen for Thing
		// (child) connections on the port range [BeginPort-EndPort].
		//
		// The bridge port range must be within the system's
		// ip_local_reserved_ports.
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
		PortBegin uint `yaml:"PortBegin"`
		// Ending port number.
		PortEnd uint `yaml:"PortEnd"`
		// Match is a regular expresion (re) to specifiy which things
		// can connect to the bridge.  The re matches against three
		// fields of the thing: ID, Model, and Name.  The re is
		// composed with these three fields seperated by ":" character:
		// "ID:Model:Name".  See
		// https://github.com/google/re2/wiki/Syntax for regular
		// expression syntax.  Examples:
		//
		//	".*:.*:.*"		Match any thing.
		//	"123456:.*:.*"		Match only a thing with ID=123456
		//	".*:chat:.*"		Match only chat things
		Match string `yaml:"Match"`
	} `yaml:"Bridge"`
}

// All things implement this interface
type Thinger interface {
	// List of subscribers on thing bus.  On packet receipt, the
	// subscribers are process in-order, and the first matching subscriber
	// stops the processing.
	Subscribers() Subscribers
	// Path to thing's home page template
	Template() string
}

type Thingers map[string]func() Thinger

// Thing's backing structure
type Thing struct {
	thinger     Thinger
	cfg         *ThingConfig
	status      string
	id          string
	model       string
	name        string
	startupTime time.Time
	bus         *bus
	tunnel      *tunnel
	private     *webPrivate
	public      *webPublic
	assetsDir   string
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
		status:      "online",
		id:          id,
		model:       cfg.Thing.Model,
		name:        cfg.Thing.Name,
		startupTime: time.Now(),
		assetsDir:   cfg.Thing.AssetsDir,
		isPrime:     isPrime,
		log:         log.New(os.Stderr, prefix, 0),
	}

	t.bus = newBus(t, 10, thinger.Subscribers())

	t.tunnel = newTunnel(t.id, cfg.Mother.Host, cfg.Mother.User,
		cfg.Mother.Key, cfg.Thing.PortPrivate, cfg.Mother.PortPrivate)

	t.private = newWebPrivate(t, cfg.Thing.PortPrivate)
	t.public = newWebPublic(t, cfg.Thing.PortPublic, cfg.Thing.PortPublicTLS,
		cfg.Thing.User)

	t.templ, t.templErr = template.ParseFiles(thinger.Template())

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

func (t *Thing) runPrime() error {
	t.private.start()
	t.public.start()
	t.tunnel.start()
	return t.primePort.run()
}

func (t *Thing) Run() error {
	switch {
	case t.isPrime:
		return t.runPrime()
	default:
		return t.run()
	}
}

// Run a copy of the thing (shadow thing) in the bridge.
func (t *Thing) runInBridge(p *port) {
	var name = fmt.Sprintf("port:%d", p.port)
	var sock = newWebSocket(name, p.ws)
	var pkt = newPacket(t.bus, sock, nil)
	var err error

	t.bus.plugin(sock)

	// Send a CmdStart message on startup of shadow thing so shadow thing
	// can get the current state from the real thing
	msg := struct{ Msg string }{Msg: "_CmdStart"}
	t.bus.receive(pkt.Marshal(&msg))

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
}
