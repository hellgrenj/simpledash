package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hellgrenj/simpledash/cluster"
	c "github.com/hellgrenj/simpledash/context"
)

type server struct {
	router            *mux.Router
	SimpledashContext c.SimpledashContext
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) serveIndexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func (s *server) getContext(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.SimpledashContext)
}

func (s *server) startListen() {
	log.Println("simpledash now running on *:1337")
	log.Fatal(http.ListenAndServe(":1337", s))
}

func newServer() *server {
	sc := c.GetContext()
	s := &server{router: mux.NewRouter(), SimpledashContext: sc}
	s.router.HandleFunc("/", s.serveIndexPage).Methods("GET")
	s.router.HandleFunc("/context", s.getContext).Methods("GET")
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	s.router.HandleFunc("/ws", connect)
	return s
}

func Serve(clusterInfoChan <-chan cluster.ClusterInfo) {
	go publishEventsToWsConnections(clusterInfoChan)
	log.Println("Starting server...")
	s := newServer()
	s.startListen()
}
