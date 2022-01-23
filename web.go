package merle

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"errors"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/msteinert/pam"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var upgrader = websocket.Upgrader{}

// Open a websocket on the thing
func (t *Thing) ws(w http.ResponseWriter, r *http.Request) {
	var err error

	vars := mux.Vars(r)
	id := vars["id"]

	// If this thing is a bridge, and the ID matches a child ID, then hand
	// the websocket request to the child.
	child := t.getChild(id)
	if child != nil {
		child.ws(w, r)
		return
	}

	if id != "" && id != t.id {
		t.log.Println("Mismatch on Ids")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		t.log.Println("Websocket upgrader error:", err)
		return
	}
	defer ws.Close()

	name := "ws:" + r.RemoteAddr + r.RequestURI
	var sock = newWebSocket(name, ws)

	t.log.Printf("Websocket opened [%s]", name)

	// Plug the websocket into the thing's bus
	t.bus.plugin(sock)

	for {
		// New pkt for each rcv
		var pkt = newPacket(t.bus, sock, nil)

		_, pkt.msg, err = ws.ReadMessage()
		if err != nil {
			t.log.Printf("Websocket closed [%s:%s]",
				r.RemoteAddr, r.RequestURI)
			break
		}

		// Put the packet on the bus
		t.bus.receive(pkt)
	}

	// Unplug the websocket from the thing's bus
	t.bus.unplug(sock)
}

// Some things to pass into the thing's HTML template
func (t *Thing) homeParams(r *http.Request) interface{} {
	scheme := "wss://"
	if r.TLS == nil {
		scheme = "ws://"
	}

	return struct {
		Scheme string
		Host   string
		Status string
		Id     string
		Model  string
		Name   string
	}{
		Scheme: scheme,
		Host:   r.Host,
		Status: t.status,
		Id:     t.id,
		Model:  t.model,
		Name:   t.name,
	}
}

// Open the thing's home page
func (t *Thing) home(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// If this thing is a bridge, and the ID matches a child ID, then open
	// the child's home page
	child := t.getChild(id)
	if child != nil {
		child.home(w, r)
		return
	}

	if id != "" && id != t.id {
		http.Error(w, "Mismatch on Ids", http.StatusNotFound)
		return
	}

	if t.templErr == nil {
		t.templ.Execute(w, t.homeParams(r))
	} else {
		http.Error(w, t.templErr.Error(), http.StatusNotFound)
	}
}

func (t *Thing) pamValidate(user, passwd string) (bool, error) {
	trans, err := pam.StartFunc("", user,
		func(s pam.Style, msg string) (string, error) {
			switch s {
			case pam.PromptEchoOff:
				return passwd, nil
			}
			return "", errors.New("Unrecognized message style")
		})
	if err != nil {
		t.log.Println("PAM Start:", err)
		return false, err
	}
	err = trans.Authenticate(0)
	if err != nil {
		t.log.Printf("Authenticate [%s,%s]: %s", user, passwd, err)
		return false, err
	}

	return true, nil
}

func (t *Thing) basicAuth(authUser string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// skip basic authentication if no user
		if authUser == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, passwd, ok := r.BasicAuth()

		if ok {
			userHash := sha256.Sum256([]byte(user))
			expectedUserHash := sha256.Sum256([]byte(authUser))

			// https://www.alexedwards.net/blog/basic-authentication-in-go
			userMatch := (subtle.ConstantTimeCompare(userHash[:],
				expectedUserHash[:]) == 1)

			// Use PAM to validate passwd
			passwdMatch, _ := t.pamValidate(user, passwd)

			if userMatch && passwdMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// The thing's private HTTP server
type webPrivate struct {
	sync.WaitGroup
	port   uint
	mux    *mux.Router
	server *http.Server
}

func newWebPrivate(t *Thing, port uint) *webPrivate {
	addr := ":" + strconv.FormatUint(uint64(port), 10)

	mux := mux.NewRouter()
	mux.HandleFunc("/ws", t.ws)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
		// TODO add timeouts
	}

	return &webPrivate{
		port:   port,
		mux:    mux,
		server: server,
	}
}

func (w *webPrivate) start() {
	if w.port == 0 {
		log.Println("Skipping private HTTP server; port is zero")
		return
	}

	w.Add(2)
	w.server.RegisterOnShutdown(w.Done)

	log.Println("Private HTTP server listening on", w.server.Addr)

	go func() {
		if err := w.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln("Private HTTP server failed:", err)
		}
		w.Done()
	}()
}

func (w *webPrivate) stop() {
	if w.port != 0 {
		w.server.Shutdown(context.Background())
	}
	w.Wait()
}

func (w *webPrivate) handleFunc(pattern string,
	handler func(http.ResponseWriter, *http.Request)) {
	w.mux.HandleFunc(pattern, handler)
}

// The thing's public HTTP server
type webPublic struct {
	thing *Thing
	sync.WaitGroup
	user      string
	port      uint
	portTLS   uint
	mux       *mux.Router
	server    *http.Server
	serverTLS *http.Server
}

func newWebPublic(t *Thing, port, portTLS uint, user string) *webPublic {
	addr := ":" + strconv.FormatUint(uint64(port), 10)
	addrTLS := ":" + strconv.FormatUint(uint64(portTLS), 10)

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("./certs"),
	}

	fs := http.FileServer(http.Dir(t.assets.Dir))

	mux := mux.NewRouter()
	mux.HandleFunc("/ws/{id}", t.basicAuth(user, t.ws))
	mux.HandleFunc("/{id}", t.basicAuth(user, t.home))
	mux.HandleFunc("/", t.basicAuth(user, t.home))
	mux.PathPrefix("/" + t.id + "/assets/").
		Handler(http.StripPrefix("/" + t.id + "/assets/", fs))

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
		// TODO add timeouts
	}

	if portTLS != 0 {
		server.Handler = certManager.HTTPHandler(nil)
	}

	serverTLS := &http.Server{
		Addr:    addrTLS,
		Handler: mux,
		// TODO add timeouts
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	return &webPublic{
		thing:     t,
		user:      user,
		port:      port,
		portTLS:   portTLS,
		mux:       mux,
		server:    server,
		serverTLS: serverTLS,
	}
}

func (w *webPublic) start() {
	if w.port == 0 {
		log.Println("Skipping public HTTP server; port is zero")
		return
	}

	if w.thing.assets.Dir == "" {
		log.Println("Skipping public HTTP server; assets directory is missing")
		return
	}

	if w.user != "" {
		log.Println("Basic Authencation enabled for user", w.user)
	}

	w.Add(2)
	w.server.RegisterOnShutdown(w.Done)

	log.Println("Public HTTP server listening on", w.server.Addr)

	go func() {
		if err := w.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln("Public HTTP server failed:", err)
		}
		w.Done()
	}()

	if w.portTLS == 0 {
		log.Println("Skipping public HTTPS server; port is zero")
		return
	}

	w.Add(2)
	w.serverTLS.RegisterOnShutdown(w.Done)

	log.Println("Public HTTPS server listening on", w.serverTLS.Addr)

	go func() {
		if err := w.serverTLS.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.Fatalln("Public HTTPS server failed:", err)
		}
		w.Done()
	}()
}

func (w *webPublic) stop() {
	if w.port != 0 {
		w.server.Shutdown(context.Background())
	}
	if w.portTLS != 0 {
		w.serverTLS.Shutdown(context.Background())
	}
	w.Wait()
}
