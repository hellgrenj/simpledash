package main

func main() {
	clusterInfoChan := make(chan ClusterInfo)
	go MonitorCluster(clusterInfoChan)
	Serve(clusterInfoChan)
}
