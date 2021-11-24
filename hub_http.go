package merle

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"html/template"
	"log"
	"net/http"
	"time"
)

func getPort(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if len(q) == 1 {
		id := q.Get("id")
		if len(id) > 0 {
			port := portFromId(id)
			switch port {
			case -1:
				fmt.Fprintf(w, "no ports available")
			case -2:
				fmt.Fprintf(w, "port busy")
			default:
				fmt.Fprintf(w, "%d", port)
			}
			return
		}
	}
	http.Error(w, "Missing device ID", http.StatusNotFound)
}

func (h *Hub) wsDevice(w http.ResponseWriter, r *http.Request, id string) {
	/*
		var s socket
		var err error

		d := Hub.getDevice(id)
		if d == nil {
			http.Error(w, "Unknown device ID "+id, http.StatusNotFound)
			return
		}

		s.conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Websocket upgrader error:", err)
			return
		}

		d.SocketAdd(&s)

		for {
			_, msg, err := s.conn.ReadMessage()
			if err != nil {
				log.Println("wsDevice read message error:", err)
				break
			}
			d.Receive(&s, msg)
		}

		d.SocketDelete(&s)
		s.conn.Close()
	*/
}

func (h *Hub) wsHub(w http.ResponseWriter, r *http.Request) {
	var p Packet
	var err error

	p.conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Websocket upgrader error:", err)
		return
	}

	h.connAdd(p.conn)

	for {
		_, p.Msg, err = p.conn.ReadMessage()
		if err != nil {
			log.Println("wsHub read message error:", err)
			break
		}
		h.receivePacket(&p)
	}

	h.connDelete(p.conn)
	p.conn.Close()
}

func (h *Hub) ws(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if len(q) == 1 {
		id := q.Get("id")
		h.wsDevice(w, r, id)
	} else {
		h.wsHub(w, r)
	}
}

func (h *Hub) homeDevice(w http.ResponseWriter, r *http.Request, id string) {
	d := h.getDevice(id)
	if d == nil {
		http.Error(w, "Unknown device ID "+id, http.StatusNotFound)
		return
	}

	d.m.HomePage(w, r)
}

var homeTempl = template.Must(template.ParseFiles("./res/templates/hub.html"))

func (h *Hub) homeHub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

func (h *Hub) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	if len(q) > 1 {
		s := fmt.Sprintf("Unexpected URL query: %s", r.URL)
		http.Error(w, s, http.StatusBadRequest)
		return
	} else if len(q) == 1 {
		id := q.Get("id")
		if len(id) > 0 {
			h.homeDevice(w, r, id)
		} else {
			s := fmt.Sprintf("Unexpected URL query: %s", r.URL)
			http.Error(w, s, http.StatusBadRequest)
			return
		}
	} else {
		h.homeHub(w, r)
	}
}

func (h *Hub) http() {
	privateMux := http.NewServeMux()
	privateMux.HandleFunc("/port", getPort)
	privateMux.HandleFunc("/ws", h.ws)

	go func() {
		log.Printf("Listening HTTP on :8080 for private\n")
		err := http.ListenAndServe(":8080", privateMux)
		log.Fatalln("Private HTTP server failed:", err)
	}()

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("./certs"),
	}

	go func() {
		log.Printf("Listening HTTP on :80 for public\n")
		err := http.ListenAndServe(":80", certManager.HTTPHandler(nil))
		log.Fatalln("Public HTTP server failed:", err)
	}()

	fs := http.FileServer(http.Dir("./res"))

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/ws", h.ws)
	publicMux.HandleFunc("/", h.home)
	publicMux.Handle("/res/", http.StripPrefix("/res", fs))

	https := &http.Server{
		Addr:         ":443",
		Handler:      publicMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	log.Printf("Listening HTTPS on %s for public\n", https.Addr)
	err := https.ListenAndServeTLS("", "")
	log.Fatalln("Public HTTPS server failed:", err)
}
