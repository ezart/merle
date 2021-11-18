package merle

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/msteinert/pam"
	"net/http"
	"log"
)

var upgrader = websocket.Upgrader{}

func (d *Device) ws(w http.ResponseWriter, r *http.Request) {
	var p Packet
	var err error

	p.conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Websocket upgrader error:", err)
		return
	}

	d.connAdd(p.conn)

	for {
		_, p.Msg, err = p.conn.ReadMessage()
		if err != nil {
			log.Println("Websocket read message error:", err)
			break
		}
		d.receivePacket(&p)
	}

	d.connDelete(p.conn)
	p.conn.Close()
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

	d.homePage(w, r)
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

func (d *Device) http(authUser string) {
	privateMux := http.NewServeMux()
	privateMux.HandleFunc("/ws", d.ws)

	go func() {
		log.Printf("Listening HTTP on :8080 for private")
		err := http.ListenAndServe(":8080", privateMux)
		log.Fatalln("Private HTTP server failed:", err)
	}()

	fs := http.FileServer(http.Dir("res"))

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/ws", basicAuth(authUser, d.ws))
	publicMux.HandleFunc("/", basicAuth(authUser, d.home))
	publicMux.Handle("/res/", http.StripPrefix("/res", fs))

	log.Printf("Listening HTTP on :80 for public")
	err := http.ListenAndServe(":80", publicMux)
	log.Fatalln("Public HTTP server failed:", err)
}
