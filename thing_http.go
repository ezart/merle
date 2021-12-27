// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
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
	"log"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{}

func (t *Thing) ws(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Println("hitting ws id", id)

	child := t.getThing(id)
	if child != nil {
		child.ws(w, r)
		return
	}

	if id != "" && id != t.id {
		log.Println(t.logPrefix(), "Mismatch on Ids")
		return
	}

	t.connQ <- true
	defer func() { <-t.connQ }()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(t.logPrefix(), "Websocket upgrader error:", err)
		return
	}
	defer conn.Close()

	t.connAdd(conn)

	for {
		var p = &Packet{
			conn: conn,
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(t.logPrefix(), "Websocket read error:", err)
			break
		}
		t.receive(p.update(msg))
	}

	t.connDelete(conn)
}

func (t *Thing) home(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Println("hitting home id", id)
	log.Println("hitting home r.URL.Path", r.URL.Path)

	/*
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	*/
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	child := t.getThing(id)
	if child != nil {
		log.Println("hitting home for child", child.id)
		child.home(w, r)
		return
	}

	log.Println("hitting home here")
	if id != "" && id != t.id {
		http.Error(w, "Mismatch on Ids", http.StatusNotFound)
		return
	}

	log.Println("hitting home here2")
	if t.Home == nil {
		http.Error(w, "Home page not set up", http.StatusNotFound)
		return
	}

	log.Println("hitting home here3")
	t.Home(w, r)
}

func pamValidate(user, passwd string) (bool, error) {
	t, err := pam.StartFunc("", user, func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return passwd, nil
		}
		return "", errors.New("Unrecognized message style")
	})
	if err != nil {
		log.Println("PAM Start:", err)
		return false, err
	}
	err = t.Authenticate(0)
	if err != nil {
		log.Printf("Authenticate [%s,%s]: %s", user, passwd, err)
		return false, err
	}

	return true, nil
}

func basicAuth(authUser string, next http.HandlerFunc) http.HandlerFunc {
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
			passwdMatch, _ := pamValidate(user, passwd)

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
	t.Lock()
	for c := range t.conns {
		c.WriteControl(websocket.CloseMessage, nil, time.Now())
	}
	t.Unlock()
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

	log.Printf("%s Private HTTP listening on %s", t.logPrefix(), addrPrivate)

	go func() {
		if err := t.httpPrivate.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln(t.logPrefix(), "Private HTTP server failed:", err)
		}
		t.Done()
	}()
}

func (t *Thing) httpInitPublic() {
	fs := http.FileServer(http.Dir("web"))
	t.muxPublic = mux.NewRouter()
	t.muxPublic.HandleFunc("/ws/{id}", basicAuth(t.authUser, t.ws))
	t.muxPublic.HandleFunc("/{id}", basicAuth(t.authUser, t.home))
	t.muxPublic.HandleFunc("/", basicAuth(t.authUser, t.home))
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

	log.Printf("%s Public HTTP listening on %s", t.logPrefix(), addrPublic)

	go func() {
		if err := t.httpPublic.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln("Public HTTP server failed:", err)
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
		log.Println(t.logPrefix(), "Skipping private HTTP")
	} else {
		t.httpStartPrivate()
	}
	if t.portPublic == 0 {
		log.Println(t.logPrefix(), "Skipping public HTTP")
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

func getPort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	port := portFromId(id)

	switch port {
	case -1:
		fmt.Fprintf(w, "no ports available")
	case -2:
		fmt.Fprintf(w, "port busy")
	default:
		fmt.Fprintf(w, "%d", port)
	}
}

func (t *Thing) ListenForThings() error {
	log.Println("Listening for Things...")
	t.HandleMsg("GetThings", t.getThings)
	t.muxPrivate.HandleFunc("/port/{id}", getPort)
	return t.portScan()
}
