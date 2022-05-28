package main

import (
	"encoding/json"
	"log"
	"os"
)

type SimpledashContext struct {
	ClusterName string
	Namespaces  []string
}

func getContext() SimpledashContext {
	clusterName := os.Getenv("SIMPLEDASH_CLUSTERNAME")
	if clusterName == "" {
		log.Println("failed to fetch SIMPLEDASH_CLUSTERNAME environment variable")
		clusterName = "unknown cluster"
	}
	var namespaces []string
	err := json.Unmarshal([]byte(os.Getenv("SIMPLEDASH_NAMESPACES")), &namespaces)
	if err != nil {
		log.Println("failed to parse SIMPLEDASH_NAMESPACES environment variable as JSON")
		log.Println("checking all namespaces...")
		namespaces = []string{""}
	}
	sc := SimpledashContext{ClusterName: clusterName, Namespaces: namespaces}
	return sc
}
