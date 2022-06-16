package server

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hellgrenj/simpledash/cluster"
	c "github.com/hellgrenj/simpledash/context"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

const (
	// Time allowed to read the next pong message from the client.
	pongWait = 10 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

func connect(hub *Hub, w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	c.SetReadLimit(512)
	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error { c.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	go func() { // write pings to client..
		pingTicker := time.NewTicker(pingPeriod)
		for range pingTicker.C {
			c.SetWriteDeadline(time.Now().Add(30 * time.Second))
			if err := c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}()
	go func() { // read from client
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				log.Printf("removing connection %v. received %v", c.RemoteAddr(), err)
				hub.unregister <- c
				break
			}
		}
	}()
	hub.register <- c
}

type server struct {
	router            *mux.Router
	SimpledashContext c.SimpledashContext
	hub               *Hub
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
	s := &server{router: mux.NewRouter(), SimpledashContext: sc, hub: newHub()}
	go s.hub.run()
	s.router.HandleFunc("/", s.serveIndexPage).Methods("GET")
	s.router.HandleFunc("/context", s.getContext).Methods("GET")
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	s.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connect(s.hub, w, r)
	})
	return s
}
func pipeToHub(clusterInfoChan <-chan cluster.ClusterInfo, hub *Hub) {
	for {
		clusterInfo := <-clusterInfoChan
		hub.clusterInfo <- clusterInfo
	}
}
func Serve(clusterInfoChan <-chan cluster.ClusterInfo) {
	log.Println("Starting server...")
	s := newServer()
	go pipeToHub(clusterInfoChan, s.hub)
	s.startListen()
}
