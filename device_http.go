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

func (d *Device) ws(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Websocket upgrader error:", err)
		return
	}

	d.connAdd(conn)

	for {
		var p = &Packet{
			conn: conn,
		}

		_, p.Msg, err = conn.ReadMessage()
		if err != nil {
			log.Println("Websocket read message error:", err)
			break
		}
		d.receive(p)
	}

	d.connDelete(conn)
	conn.Close()
}

func (d *Device) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	d.m.HomePage(w, r)
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

func (d *Device) shutdown() {
	d.Lock()
	for c := range d.conns {
		c.WriteControl(websocket.CloseMessage, nil, time.Now())
	}
	d.Unlock()
	d.wg.Done()
}

func (d *Device) httpStop() {
	d.privateServer.Shutdown(context.Background())
	d.publicServer.Shutdown(context.Background())
	d.wg.Wait()
}

func (d *Device) httpStart(authUser string, publicPort, privatePort int) {
	publicAddr := ":" + strconv.Itoa(publicPort)
	privateAddr := ":" + strconv.Itoa(privatePort)

	fs := http.FileServer(http.Dir("web"))

	privateMux := http.NewServeMux()
	privateMux.HandleFunc("/ws", d.ws)

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/ws", basicAuth(authUser, d.ws))
	publicMux.HandleFunc("/", basicAuth(authUser, d.home))
	publicMux.Handle("/web/", http.StripPrefix("/web", fs))

	d.privateServer = &http.Server{
		Addr:    privateAddr,
		Handler: privateMux,
	}

	d.publicServer = &http.Server{
		Addr:    publicAddr,
		Handler: publicMux,
	}

	d.wg.Add(2)
	d.privateServer.RegisterOnShutdown(d.shutdown)

	go func() {
		log.Printf("Listening HTTP on %s for private", privateAddr)
		if err := d.privateServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln("Private HTTP server failed:", err)
		}
		d.wg.Done()
	}()

	if authUser == "" {
		log.Printf("Missing authUser; skipping HTTP on %s for public", publicAddr)
		return
	}

	d.wg.Add(2)
	d.publicServer.RegisterOnShutdown(d.shutdown)

	go func() {
		log.Printf("Listening HTTP on %s for public", publicAddr)
		if err := d.publicServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln("Public HTTP server failed:", err)
		}
		d.wg.Done()
	}()
}
