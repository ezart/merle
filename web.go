package merle

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/msteinert/pam"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type weber interface {
	Start()
	Stop()
	HandleFunc(string, func(http.ResponseWriter, *http.Request))
}

var upgrader = websocket.Upgrader{}

func (t *Thing) ws(w http.ResponseWriter, r *http.Request) {
	var err error

	vars := mux.Vars(r)
	id := vars["id"]

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

	t.bus.plugin(sock)

	for {
		// new pkt for each rcv
		var pkt = newPacket(t.bus, sock, nil)

		_, pkt.msg, err = ws.ReadMessage()
		if err != nil {
			t.log.Printf("Websocket closed [%s:%s]",
				r.RemoteAddr, r.RequestURI)
			break
		}

		t.bus.receive(pkt)
	}

	t.bus.unplug(sock)
}

func (t *Thing) HomeParams(r *http.Request) interface{} {
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

func (t *Thing) home(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	child := t.getChild(id)
	if child != nil {
		child.home(w, r)
		return
	}

	if id != "" && id != t.Id() {
		http.Error(w, "Mismatch on Ids", http.StatusNotFound)
		return
	}

	if t.templErr == nil {
		t.templ.Execute(w, t.HomeParams(r))
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

type webPrivate struct {
	sync.WaitGroup
	port   uint
	mux    *mux.Router
	server *http.Server
}

func newWebPrivate(t *Thing, port uint) weber {
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

func (w *webPrivate) Start() {
	if w.port == 0 {
		log.Println("Skipping private HTTP server")
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

func (w *webPrivate) Stop() {
	if w.port != 0 {
		w.server.Shutdown(context.Background())
	}
	w.Wait()
}

func (w *webPrivate) HandleFunc(pattern string,
	handler func(http.ResponseWriter, *http.Request)) {
	w.mux.HandleFunc(pattern, handler)
}

type webPublic struct {
	sync.WaitGroup
	user   string
	port   uint
	server *http.Server
}

func newWebPublic(t *Thing, user string, port uint) weber {
	addr := ":" + strconv.FormatUint(uint64(port), 10)

	fs := http.FileServer(http.Dir("web"))

	mux := mux.NewRouter()
	mux.HandleFunc("/ws/{id}", t.basicAuth(user, t.ws))
	mux.HandleFunc("/{id}", t.basicAuth(user, t.home))
	mux.HandleFunc("/", t.basicAuth(user, t.home))
	mux.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
		// TODO add timeouts
	}

	return &webPublic{
		user:   user,
		port:   port,
		server: server,
	}
}

func (w *webPublic) Start() {
	if w.port == 0 {
		log.Println("Skipping public HTTP server")
		return
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
}

func (w *webPublic) Stop() {
	if w.port != 0 {
		w.server.Shutdown(context.Background())
	}
	w.Wait()
}

func (w *webPublic) HandleFunc(pattern string,
	handler func(http.ResponseWriter, *http.Request)) {
}
