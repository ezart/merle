package merle

import (
	"fmt"
	"log"
	"net/http"
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

func (h *Hub) http() {
	privateMux := http.NewServeMux()
	privateMux.HandleFunc("/port", getPort)

	log.Printf("Listening HTTP on :8080 for private\n")
	err := http.ListenAndServe(":8080", privateMux)
	log.Fatalln("Private HTTP server failed:", err)
}
