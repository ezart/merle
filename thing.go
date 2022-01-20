package merle

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"time"
)

// Thing configuration.  All Things share this configuration.
type ThingConfig struct {

	// The section describes a Thing.
	Thing struct {
		// (Optional) Thing's Id.  Ids are unique within an application
		// to differenciate one Thing from another.  Id is optional; if
		// Id is not given, a system-wide unique Id is assigned.
		Id string `yaml:"Id"`
		// Thing's Model.  Should match one of the models support by
		// Merle.  See merle --models for list of support models.
		Model string `yaml:"Model"`
		// Thing's Name
		Name string `yaml:"Name"`
		// (Optional) system User.  If a User is given, any browser
		// views of the Thing's home page  will prompt for user/passwd.
		// HTTP Basic Authentication is used and the user/passwd given
		// must match the system creditials for the User.  If no User
		// is given, HTTP Basic Authentication is skipped; anyone can
		// view the home page.
		User string `yaml:"User"`
		// (Optional) If PortPublic is given, an HTTP web server is
		// started on port PortPublic.  PortPublic is typically set to
		// 80.  The HTTP web server runs the Thing's home page.
		PortPublic uint `yaml:"PortPublic"`
		// (Optional) If PortPublicTLS is given, an HTTPS web server is
		// started on port PortPublicTLS.  PortPublicTLS is typically
		// set to 443.  The HTTPS web server will self-certify using a
		// certificate from Let's Encrypt.  The public HTTPS server
		// will securely run the Thing's home page.  If PortPublicTLS
		// is given, PortPublic must be given.
		PortPublicTLS uint `yaml:"PortPublicTLS"`
		// (Optional) If PortPrivate is given, a HTTP server is
		// started on port PortPrivate.  This HTTP server does not
		// server up the Thing's home page but rather connects to
		// Thing's Mother using a websocket over HTTP.
		PortPrivate uint `yaml:"PortPrivate"`
	} `yaml:"Thing"`

	// (Optional) This section describes a Thing's Mother.  Every Thing has
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
}

// All things implement this interface
type Thinger interface {
	// List of subscribers on thing bus.  On packet receipt, the
	// subscribers are process in-order, and the first matching subscriber
	// stops the processing.
	Subscribe() Subscribers
	// Thing configurator
	Config(Configurator) error
	// Path to thing's home page template
	Template() string
}

// Thing's backing structure
type thing struct {
	thinger     Thinger
	status      string
	id          string
	model       string
	name        string
	startupTime time.Time
	config      Configurator
	bus         *bus
	tunnel      *tunnel
	private     *webPrivate
	public      *webPublic
	templ       *template.Template
	templErr    error
	isBridge    bool
	bridge      *bridge
	log         *log.Logger
}

func NewThing(thinger Thinger, id, model, name string) *thing {
	id = defaultId(id)

	prefix := "[" + id + "] "
	log := log.New(os.Stderr, prefix, 0)

	t := &thing{
		thinger:     thinger,
		status:      "online",
		id:          id,
		model:       model,
		name:        name,
		bus:         newBus(log, 10, thinger.Subscribe()),
		log:         log,
	}

	t.bus.subscribe("_GetIdentity", t.getIdentity)

	return t
}

func (t *thing) EnablePublicHTTP(port, portTLS uint, user, assetsDir string) {
	t.public = newWebPublic(t, port, portTLS, user, assetsDir)
}

func (t *thing) EnablePrivateHTTP(port uint) {
	t.private = newWebPrivate(t, port)
}

func (t *thing) SetTemplate(file string) {
	t.templ, t.templErr = template.ParseFiles(file)
}

func newThing(stork Storker, config Configurator, demo bool) (*thing, error) {
	var cfg ThingConfig
	var thinger Thinger
	var err error

	if err = config.Parse(&cfg); err != nil {
		return nil, err
	}

	id := defaultId(cfg.Thing.Id)

	prefix := "[" + id + "] "
	log := log.New(os.Stderr, prefix, 0)

	thinger, err = stork.NewThinger(log, cfg.Thing.Model, demo)
	if err != nil {
		return nil, err
	}

	t := &thing{
		thinger:     thinger,
		status:      "online",
		id:          id,
		model:       cfg.Thing.Model,
		name:        cfg.Thing.Name,
		startupTime: time.Now(),
		config:      config,
		bus:         newBus(log, 10, thinger.Subscribe()),
		log:         log,
	}

	t.tunnel = newTunnel(t.id, cfg.Mother.Host, cfg.Mother.User,
		cfg.Mother.Key, cfg.Thing.PortPrivate, cfg.Mother.PortPrivate)

	t.private = newWebPrivate(t, cfg.Thing.PortPrivate)
	t.public = newWebPublic(t, cfg.Thing.PortPublic, cfg.Thing.PortPublicTLS,
		cfg.Thing.User, "web")

	t.templ, t.templErr = template.ParseFiles(thinger.Template())

	_, t.isBridge = t.thinger.(Bridger)
	if t.isBridge {
		t.bridge, err = newBridge(log, stork, config, t)
		if err != nil {
			return nil, err
		}
	}

	t.bus.subscribe("GetIdentity", t.getIdentity)

	return t, err
}

// RunThing is the main entry point for Merle.  A new Thing is created and run.
//
// The stork delivers a new Thing based on the config.  The config names the
// Thing and sets the Thing's properties.  Any error creating or running the
// Thing is returned.  The Thing should run forever.  It is an error for a
// Thing to stop running once started.
//
// Demo is set true to run Thing in demo-mode.  In demo-mode, the Thing will
// similuate hardware access.  Demo-mode is handy for testing functionality
// without having access to hardware.
func RunThing(stork Storker, config Configurator, demo bool) error {

	thing, err := newThing(stork, config, demo)
	if err != nil {
		return err
	}

	return thing.run()
}

func (t *thing) Run() error {
	t.startupTime = time.Now()
	return t.run()
}

type msgIdentity struct {
	Msg         string
	Status      string
	Id          string
	Model       string
	Name        string
	StartupTime time.Time
}

func (t *thing) getIdentity(p *Packet) {
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

func (t *thing) getChild(id string) *thing {
	if !t.isBridge {
		return nil
	}
	return t.bridge.getChild(id)
}

func (t *thing) privateStart() {
	if (t.private != nil) {
		t.private.start()
	}
}

func (t *thing) privateStop() {
	if (t.private != nil) {
		t.private.stop()
	}
}

func (t *thing) publicStart() {
	if (t.public != nil) {
		t.public.start()
	}
}

func (t *thing) publicStop() {
	if (t.public != nil) {
		t.public.stop()
	}
}

func (t *thing) tunnelStart() {
	if (t.tunnel != nil) {
		t.tunnel.start()
	}
}

func (t *thing) tunnelStop() {
	if (t.tunnel != nil) {
		t.tunnel.stop()
	}
}

func (t *thing) run() error {

	t.privateStart()
	t.publicStart()
	t.tunnelStart()

	if t.isBridge {
		t.bridge.Start()
	}

	msg := struct{ Msg string }{Msg: "_CmdRun"}
	t.bus.receive(newPacket(t.bus, nil, &msg))

	if t.isBridge {
		t.bridge.Stop()
	}

	t.tunnelStop()
	t.publicStop()
	t.privateStop()

	t.bus.close()

	return fmt.Errorf("_CmdRun didn't run forever")
}

// Run a copy of the thing (shadow thing) in the bridge.
func (t *thing) runInBridge(p *port) {
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
