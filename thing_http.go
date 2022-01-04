// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/msteinert/pam"
	"net/http"
	"strconv"
)

var upgrader = websocket.Upgrader{}

func (t *Thing) ws(w http.ResponseWriter, r *http.Request) {
	var err error

	vars := mux.Vars(r)
	id := vars["id"]

	child := t.GetChild(id)
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
		var pkt = newPacket(sock, nil)

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

func (t *Thing) home(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	child := t.GetChild(id)
	if child != nil {
		child.home(w, r)
		return
	}

	if id != "" && id != t.id {
		http.Error(w, "Mismatch on Ids", http.StatusNotFound)
		return
	}

	if t.Home == nil {
		http.Error(w, "Home page not set up", http.StatusNotFound)
		return
	}

	t.Home(w, r)
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

		if authUser == "testtest" {
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

func (t *Thing) httpShutdown() {
	t.bus.close()
	t.Done()
}

func (t *Thing) httpInitPrivate() {
	t.muxPrivate = mux.NewRouter()
	t.muxPrivate.HandleFunc("/ws", t.ws)
}

func (t *Thing) httpStartPrivate() {
	addrPrivate := ":" + strconv.Itoa(t.portPrivate)

	t.httpPrivate = &http.Server{
		Addr:    addrPrivate,
		Handler: t.muxPrivate,
		// TODO add timeouts
	}

	t.Add(2)
	t.httpPrivate.RegisterOnShutdown(t.httpShutdown)

	t.log.Println("Private HTTP listening on", addrPrivate)

	go func() {
		if err := t.httpPrivate.ListenAndServe(); err != http.ErrServerClosed {
			t.log.Fatalln("Private HTTP server failed:", err)
		}
		t.Done()
	}()
}

func (t *Thing) httpInitPublic() {
	fs := http.FileServer(http.Dir("web"))
	t.muxPublic = mux.NewRouter()
	t.muxPublic.HandleFunc("/ws/{id}", t.basicAuth(t.authUser, t.ws))
	t.muxPublic.HandleFunc("/{id}", t.basicAuth(t.authUser, t.home))
	t.muxPublic.HandleFunc("/", t.basicAuth(t.authUser, t.home))
	t.muxPublic.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))
}

func (t *Thing) httpStartPublic() {
	addrPublic := ":" + strconv.Itoa(t.portPublic)

	t.httpPublic = &http.Server{
		Addr:    addrPublic,
		Handler: t.muxPublic,
		// TODO add timeouts
	}

	t.Add(2)
	t.httpPublic.RegisterOnShutdown(t.httpShutdown)

	t.log.Println("Public HTTP listening on", addrPublic)

	go func() {
		if err := t.httpPublic.ListenAndServe(); err != http.ErrServerClosed {
			t.log.Fatalln("Public HTTP server failed:", err)
		}
		t.Done()
	}()
}

func (t *Thing) httpInit() {
	t.httpInitPrivate()
	t.httpInitPublic()
}

func (t *Thing) httpStart() {
	if t.portPrivate == 0 {
		t.log.Println("Skipping private HTTP")
	} else {
		t.httpStartPrivate()
	}
	if t.portPublic == 0 {
		t.log.Println("Skipping public HTTP")
	} else {
		t.httpStartPublic()
	}
}

func (t *Thing) httpStop() {
	if t.portPrivate != 0 {
		t.httpPrivate.Shutdown(context.Background())
	}
	if t.portPublic != 0 {
		t.httpPublic.Shutdown(context.Background())
	}
	t.Wait()
}

func (t *Thing) getPort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	port := t.portFromId(id)

	switch port {
	case -1:
		fmt.Fprintf(w, "no ports available")
	case -2:
		fmt.Fprintf(w, "port busy")
	default:
		fmt.Fprintf(w, "%d", port)
	}
}

// Configure local http server
func (t *Thing) HttpConfig(authUser string, portPublic, portPrivate int) {
	t.authUser = authUser
	t.portPublic = portPublic
	t.portPrivate = portPrivate
}
