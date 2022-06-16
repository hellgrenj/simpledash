package server

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/hellgrenj/simpledash/cluster"
)

type Hub struct {
	connections map[*websocket.Conn]bool
	clusterInfo chan cluster.ClusterInfo
	register    chan *websocket.Conn
	unregister  chan *websocket.Conn
	latestInfo  *cluster.ClusterInfo
}

func newHub() *Hub {
	return &Hub{
		clusterInfo: make(chan cluster.ClusterInfo),
		register:    make(chan *websocket.Conn),
		unregister:  make(chan *websocket.Conn),
		connections: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case connection := <-h.register:
			h.connections[connection] = true
			if h.latestInfo != nil {
				log.Println("flushing latest info to new connection")
				connection.WriteJSON(h.latestInfo)
			}
			log.Printf("number of connections %v", len(h.connections))
		case connection := <-h.unregister:
			delete(h.connections, connection)
			connection.Close()
			log.Println("removed connection")
			log.Printf("number of connections %v", len(h.connections))
		case clusterInfo := <-h.clusterInfo:
			h.latestInfo = &clusterInfo
			for c := range h.connections {
				err := c.WriteJSON(clusterInfo)
				if err != nil {
					log.Printf("Error writing clusterInfo to websocket: %v", err)
				}
			}
		}
	}
}
