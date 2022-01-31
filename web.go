// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// +build !tinygo

package merle

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/msteinert/pam"
	"golang.org/x/crypto/acme/autocert"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"
)

type ThingAssets struct {

	// Directory on file system for Thing's assets (html, css, js, etc)
	// This is an absolute or relative directory.  If relative, it's
	// relative to the Thing's executable.
	Dir string

	// Directory to Thing's HTML template file, relative to
	// ThingAssets.Dir.
	Template string

	// TemplateText is text passed in lieu of a template file.
	// TemplateText takes priority over Template, if both are present.
	TemplateText string
}

type Weber interface {
	Assets() *ThingAssets
}

type web struct {
	public  *webPublic
	private *webPrivate
}

func newWeb(t *Thing, portPublic, portPublicTLS, portPrivate uint,
	user string) *web {
	return &web{
		public: newWebPublic(t, portPublic, portPublicTLS, user),
		private: newWebPrivate(t, portPrivate),
	}
}

func (w *web) start() {
	w.public.start()
	w.private.start()
}

func (w *web) stop() {
	w.private.stop()
	w.public.stop()
}

func (w *web) handlePrimePortId() {
	w.private.mux.HandleFunc("/port/{id}", w.private.getPrimePort)
}

func (w *web) handleBridgePortId() {
	w.private.mux.HandleFunc("/port/{id}", w.private.getBridgePort)
}

func (w *web) staticFiles(t *Thing) {
	if t.isWeber {
		assets := t.thinger.(Weber).Assets()
		fs := http.FileServer(http.Dir(assets.Dir))
		path := "/" + t.id + "/assets/"
		w.public.mux.PathPrefix(path).Handler(http.StripPrefix(path, fs))
	}
}

var upgrader = websocket.Upgrader{}

func (t *Thing) runOnPort(p *port) error {
	var name = fmt.Sprintf("port:%d", p.port)
	var sock = newWebSocket(name, p.ws)
	var pkt = newPacket(t.bus, sock, nil)
	var err error

	t.log.Printf("Websocket opened [%s]", name)

	t.bus.plugin(sock)

	msg := struct{ Msg string }{Msg: GetState}
	t.log.Println("Sending:", msg)
	sock.Send(pkt.Marshal(&msg))
//	t.bus.receive(pkt.Marshal(&msg))

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
			t.log.Printf("Websocket closed [%s]", name)
			break
		}

		// Put the packet on the bus
		t.bus.receive(pkt)
	}

	// Unplug the websocket from the thing's bus
	t.bus.unplug(sock)
}

// Some things to pass into the Thing's HTML template
func (t *Thing) templateParams(r *http.Request) map[string]interface{} {
	scheme := "wss://"
	if r.TLS == nil {
		scheme = "ws://"
	}

	return map[string]interface{}{
		"Host":      r.Host,
		"Status":    t.status,
		"Id":        t.id,
		"Model":     t.model,
		"Name":      t.name,
		// TODO The forward slashes are getting escaped in the output
		// TODO within <script></script> tags.  So "/" turns into "\/".
		// TODO Need to figure out why it's doing that or decide if it matters.
		"AssetsDir": template.JSStr(t.id + "/assets"),
		"WebSocket": template.JSStr(scheme + r.Host + "/ws/" + t.id),
	}
}

// Open the Thing's home page
func (t *Thing) home(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// If this Thing is a Bridge, and the ID matches a child ID, then open
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

	if t.web.public.templErr == nil {
		t.web.public.templ.Execute(w, t.templateParams(r))
	} else {
		http.Error(w, t.web.public.templErr.Error(), http.StatusNotFound)
	}
}

func (w *webPublic) pamValidate(user, passwd string) (bool, error) {
	trans, err := pam.StartFunc("", user,
		func(s pam.Style, msg string) (string, error) {
			switch s {
			case pam.PromptEchoOff:
				return passwd, nil
			}
			return "", errors.New("Unrecognized message style")
		})
	if err != nil {
		w.thing.log.Println("PAM Start:", err)
		return false, err
	}
	err = trans.Authenticate(0)
	if err != nil {
		w.thing.log.Printf("Authenticate [%s,%s]: %s", user, passwd, err)
		return false, err
	}

	return true, nil
}

func (w *webPublic) basicAuth(authUser string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {

		// skip basic authentication if no user
		if authUser == "" {
			next.ServeHTTP(writer, r)
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
			passwdMatch, _ := w.pamValidate(user, passwd)

			if userMatch && passwdMatch {
				next.ServeHTTP(writer, r)
				return
			}
		}

		writer.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
	})
}

// The Thing's public HTTP server
type webPublic struct {
	thing *Thing
	sync.WaitGroup
	user      string
	port      uint
	portTLS   uint
	mux       *mux.Router
	server    *http.Server
	serverTLS *http.Server
	templ     *template.Template
	templErr  error
}

func newWebPublic(t *Thing, port, portTLS uint, user string) *webPublic {
	addr := ":" + strconv.FormatUint(uint64(port), 10)
	addrTLS := ":" + strconv.FormatUint(uint64(portTLS), 10)

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("./certs"),
	}

	mux := mux.NewRouter()

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

	w := &webPublic{
		thing:     t,
		user:      user,
		port:      port,
		portTLS:   portTLS,
		mux:       mux,
		server:    server,
		serverTLS: serverTLS,
	}

	t.assets = t.thinger.(Weber).Assets()
	file := path.Join(t.assets.Dir, t.assets.Template)
	w.templ, w.templErr = template.ParseFiles(file)
	if t.assets.TemplateText != "" {
		w.templ, w.templErr = template.New("").Parse(t.assets.TemplateText)
	}

	mux.HandleFunc("/ws/{id}", w.basicAuth(user, t.ws))
	mux.HandleFunc("/{id}", w.basicAuth(user, t.home))
	mux.HandleFunc("/", w.basicAuth(user, t.home))

	return w
}

func (w *webPublic) start() {
	if w.port == 0 {
		log.Println("Skipping public HTTP server; port is zero")
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

// The Thing's private HTTP server
type webPrivate struct {
	thing *Thing
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
		thing:  t,
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

func (w *webPrivate) getPrimePort(writer http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	fmt.Fprintf(writer, w.thing.getPrimePort(id))
}

func (w *webPrivate) getBridgePort(writer http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	port := w.thing.bridge.ports.getPort(id)

	switch port {
	case -1:
		fmt.Fprintf(writer, "no ports available")
	case -2:
		fmt.Fprintf(writer, "port busy")
	default:
		fmt.Fprintf(writer, "%d", port)
	}
}

type webSocket struct {
	conn  *websocket.Conn
	name  string
	flags uint32
}

func newWebSocket(name string, conn *websocket.Conn) *webSocket {
	return &webSocket{name: name, conn: conn}
}

func (ws *webSocket) Send(p *Packet) error {
	return ws.conn.WriteMessage(websocket.TextMessage, p.msg)
}

func (ws *webSocket) Close() {
	ws.conn.WriteControl(websocket.CloseMessage, nil, time.Now())
}

func (ws *webSocket) Name() string {
	return ws.name
}

func (ws *webSocket) Flags() uint32 {
	return ws.flags
}

func (ws *webSocket) SetFlags(flags uint32) {
	ws.flags = flags
}
