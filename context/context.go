package context

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

type SimpledashContext struct {
	ClusterName                    string
	Namespaces                     []string
	PodLogsLinkEnabled             bool
	PodLogsLink                    string
	DeploymentLogsLinkEnabled      bool
	DeploymentLogsLink             string
	ScanIntervalInSecondsInSeconds int
}

func GetContext() SimpledashContext {
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
	podLogsLinkEnabled := os.Getenv("POD_LOGS_LINK_ENABLED") == "true"
	podLogsLink := os.Getenv("POD_LOGS_LINK")
	deploymentLogsLinkEnabled := os.Getenv("DEPLOYMENT_LOGS_LINK_ENABLED") == "true"
	deploymentLogsLink := os.Getenv("DEPLOYMENT_LOGS_LINK")

	var scanIntervalInSeconds int
	scanIntervalInSecondsStr := os.Getenv("SCAN_INTERVAL_IN_SECONDS")
	if scanIntervalInSecondsStr == "" {
		log.Println("failed to fetch SCAN_INTERVAL_IN_SECONDS environment variable, defaulting to 10")
		scanIntervalInSeconds = 10
	} else {
		scanIntervalInSeconds, err = strconv.Atoi(scanIntervalInSecondsStr)
		if err != nil {
			log.Println("failed to parse SCAN_INTERVAL_IN_SECONDS environment variable as int, defaulting to 10")
			scanIntervalInSeconds = 10
		}
	}

	sc := SimpledashContext{
		ClusterName:                    clusterName,
		Namespaces:                     namespaces,
		PodLogsLinkEnabled:             podLogsLinkEnabled,
		PodLogsLink:                    podLogsLink,
		DeploymentLogsLinkEnabled:      deploymentLogsLinkEnabled,
		DeploymentLogsLink:             deploymentLogsLink,
		ScanIntervalInSecondsInSeconds: scanIntervalInSeconds}
	return sc
}
