package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type server struct {
	router            *mux.Router
	SimpledashContext SimpledashContext
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) serveFiles(w http.ResponseWriter, r *http.Request) {
	p := "." + r.URL.Path
	if p == "./" {
		p = "./static/index.html"
	}
	http.ServeFile(w, r, p)
}
func (s *server) startListen() {
	log.Println("simpledash now running on *:1337")
	log.Fatal(http.ListenAndServe(":1337", s))
}

func (s *server) getContext(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.SimpledashContext)
}

func newServer() *server {
	sc := getContext()
	s := &server{router: mux.NewRouter(), SimpledashContext: sc}
	s.router.HandleFunc("/", s.serveFiles).Methods("GET")
	s.router.HandleFunc("/context", s.getContext).Methods("GET")
	s.router.HandleFunc("/static/app.js", s.serveFiles).Methods("GET")
	s.router.HandleFunc("/static/ws.js", s.serveFiles).Methods("GET")
	s.router.HandleFunc("/static/style.css", s.serveFiles).Methods("GET")
	s.router.HandleFunc("/ws", connect)
	return s
}

func Serve(clusterInfoChan <-chan ClusterInfo) {
	go publishEventsToWsConnections(clusterInfoChan)
	log.Println("Starting server...")
	s := newServer()
	s.startListen()
}
