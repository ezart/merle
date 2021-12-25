// Copyright 2021 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package merle

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/msteinert/pam"
	"log"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{}

func (t *Thing) ws(w http.ResponseWriter, r *http.Request) {
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

		_, p.Msg, err = conn.ReadMessage()
		if err != nil {
			log.Println(t.logPrefix(), "Websocket read error:", err)
			break
		}
		t.receive(p)
	}

	t.connDelete(conn)
}

func (t *Thing) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if t.Home == nil {
		http.Error(w, "Home page not set up", http.StatusNotFound)
		return
	}

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

func (t *Thing) httpStop() {
	if t.portPrivate != 0 {
		t.httpPrivate.Shutdown(context.Background())
	}
	if t.portPublic != 0 {
		t.httpPublic.Shutdown(context.Background())
	}
	t.Wait()
}

func (t *Thing) httpStartPrivate() {
	if t.portPrivate == 0 {
		log.Println(t.logPrefix(), "Skipping private HTTP")
		return
	}

	addrPrivate := ":" + strconv.Itoa(t.portPrivate)

	muxPrivate := http.NewServeMux()
	muxPrivate.HandleFunc("/ws", t.ws)

	t.httpPrivate= &http.Server{
		Addr:    addrPrivate,
		Handler: muxPrivate,
	}

	t.Add(2)
	t.httpPrivate.RegisterOnShutdown(t.httpShutdown)

	go func() {
		log.Printf("%s Private HTTP listening on %s", t.logPrefix(), addrPrivate)
		if err := t.httpPrivate.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln(t.logPrefix(), "Private HTTP server failed:", err)
		}
		t.Done()
	}()
}

func (t *Thing) httpStartPublic() {
	if t.portPublic == 0 || t.authUser == ""{
		log.Println(t.logPrefix(), "Skipping public HTTP")
		return
	}

	addrPublic := ":" + strconv.Itoa(t.portPublic)

	fs := http.FileServer(http.Dir("web"))

	muxPublic := http.NewServeMux()
	muxPublic.HandleFunc("/ws", basicAuth(t.authUser, t.ws))
	muxPublic.HandleFunc("/", basicAuth(t.authUser, t.home))
	muxPublic.Handle("/web/", http.StripPrefix("/web", fs))

	t.httpPublic = &http.Server{
		Addr:    addrPublic,
		Handler: muxPublic,
	}

	t.Add(2)
	t.httpPublic.RegisterOnShutdown(t.httpShutdown)

	go func() {
		log.Printf("%s Public HTTP listening on %s", t.logPrefix(), addrPublic)
		if err := t.httpPublic.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln("Public HTTP server failed:", err)
		}
		t.Done()
	}()
}

func (t *Thing) httpStart() {
	t.httpStartPrivate()
	t.httpStartPublic()
}
