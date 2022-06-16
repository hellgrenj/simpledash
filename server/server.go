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

func newServer(hub *Hub) *server {
	sc := c.GetContext()
	s := &server{router: mux.NewRouter(), SimpledashContext: sc}
	s.router.HandleFunc("/", s.serveIndexPage).Methods("GET")
	s.router.HandleFunc("/context", s.getContext).Methods("GET")
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	s.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connect(hub, w, r)
	})
	return s
}
func publishEventsToWsConnections(clusterInfoChan <-chan cluster.ClusterInfo, hub *Hub) {
	for {
		clusterInfo := <-clusterInfoChan
		hub.clusterInfo <- clusterInfo
	}
}

func Serve(clusterInfoChan <-chan cluster.ClusterInfo) {
	hub := newHub()
	go hub.run()
	go publishEventsToWsConnections(clusterInfoChan, hub)
	log.Println("Starting server...")
	s := newServer(hub)
	s.startListen()
}
