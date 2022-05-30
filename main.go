package main

import (
	"github.com/hellgrenj/simpledash/cluster"
	"github.com/hellgrenj/simpledash/server"
)

func main() {
	clusterInfoChan := make(chan cluster.ClusterInfo)
	go cluster.StartMonitor(clusterInfoChan)
	server.Serve(clusterInfoChan)
}
