package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	c "github.com/hellgrenj/simpledash/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func StartMonitor(clusterInfoChan chan<- ClusterInfo) {
	clientset := connectk8s()
	sc := c.GetContext()

	for {
		clusterInfo := scan(clientset, sc)
		clusterInfoChan <- clusterInfo
		time.Sleep(time.Second * time.Duration(sc.ScanIntervalInSecondsInSeconds))
	}
}
func getCurrentTimeAndLocation() (time.Time, *time.Location) {
	now := time.Now()
	timeZone := os.Getenv("TIMEZONE")
	if timeZone == "" {
		log.Println("failed to read TIMEZONE environment variable, using default timezone Europe/Stockholm")
		timeZone = "Europe/Stockholm"
	}
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		log.Printf("failed to load timezone %s, using default timezone Europe/Stockholm", timeZone)
		loc, _ = time.LoadLocation("Europe/Stockholm")
	}
	return now, loc
}
func scan(clientset *kubernetes.Clientset, sc c.SimpledashContext) ClusterInfo {
	now, loc := getCurrentTimeAndLocation()
	mainClusterInfo := ClusterInfo{
		Nodes:     make(NodeInfo),
		Timestamp: now.In(loc).Format("15:04:05"),
	}

	type ClusterInfoPerNamespace struct {
		clusterInfo ClusterInfo
		namespace   string
	}
	// fetch cluster info per namespace in parallel
	var chans = make([]chan ClusterInfoPerNamespace, len(sc.Namespaces))
	for i, namespace := range sc.Namespaces {
		clusterInfo := ClusterInfo{
			Nodes: make(NodeInfo),
		}
		chans[i] = make(chan ClusterInfoPerNamespace)
		// fire off goroutine to fetch cluster info for this namespace
		go func(clusterInfoChan chan ClusterInfoPerNamespace, namespace string, clusterInfo ClusterInfo) {
			// add pods, ingresses and deployments to clusterInfo serially in this goroutine for this namespace
			addPodsInfo(clientset, &clusterInfo, namespace)
			addIngressInfo(clientset, &clusterInfo, namespace)
			addDeploymentsInfo(clientset, &clusterInfo, namespace)

			infoPerNamespace := ClusterInfoPerNamespace{clusterInfo: clusterInfo, namespace: namespace}
			clusterInfoChan <- infoPerNamespace
		}(chans[i], namespace, clusterInfo)
	}

	// merge all ClusterInfoPerNamespace into one mainClusterInfo as they arrive
	numberOfMergedClusterInfoPerNamespace := 0
	for _, namespaceInfoChan := range chans {
		nsi := <-namespaceInfoChan
		clusterInfo := nsi.clusterInfo
		namespace := nsi.namespace
		log.Printf("merging clusterInfo for namespace %s into mainClusterInfo payload\n", namespace)
		// pods
		for node, pods := range clusterInfo.Nodes {
			mainClusterInfo.Nodes[node] = append(mainClusterInfo.Nodes[node], pods...)
		}
		// ingresses
		mainClusterInfo.Ingresses = append(mainClusterInfo.Ingresses, clusterInfo.Ingresses...)
		// deployments
		mainClusterInfo.Deployments = append(mainClusterInfo.Deployments, clusterInfo.Deployments...)

		numberOfMergedClusterInfoPerNamespace++
	}
	log.Printf("SCAN complete. Merged %d clusterInfo per namespace into mainClusterInfo payload\n", numberOfMergedClusterInfoPerNamespace)

	return mainClusterInfo
}

func addPodsInfo(clientset *kubernetes.Clientset, clusterInfo *ClusterInfo, namespace string) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	for _, pod := range pods.Items {
		if len(pod.Spec.Containers) == 0 {
			log.Printf("Pod %s/%s has no containers. Skipping.", pod.Namespace, pod.Name)
			continue
		}

		status := string(pod.Status.Phase) // will be running if any container in pod is running
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil { // if any container in pod is waiting, use that status
				status = containerStatus.State.Waiting.Reason
				break
			}
		}

		podInfo := PodInfo{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Image:     pod.Spec.Containers[0].Image,
			Status:    status,
		}
		clusterInfo.Nodes[pod.Spec.NodeName] = append(clusterInfo.Nodes[pod.Spec.NodeName], podInfo)
	}
}

func addDeploymentsInfo(clientset *kubernetes.Clientset, clusterInfo *ClusterInfo, namespace string) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	for d := range deployments.Items {
		deploymentInfo := DeploymentInfo{
			Namespace:     deployments.Items[d].Namespace,
			Name:          deployments.Items[d].Name,
			Replicas:      deployments.Items[d].Status.Replicas,
			ReadyReplicas: deployments.Items[d].Status.ReadyReplicas,
		}
		clusterInfo.Deployments = append(clusterInfo.Deployments, deploymentInfo)
	}
}
func addIngressInfo(clientset *kubernetes.Clientset, clusterInfo *ClusterInfo, namespace string) {
	result := clientset.CoreV1().RESTClient().Get().AbsPath(fmt.Sprintf("/apis/networking.k8s.io/v1/namespaces/%s/ingresses", namespace)).Do(context.TODO())
	ingressInfo, err := result.Raw()
	if err != nil {
		log.Println(err)
		return
	}
	var ingress Ingress
	err = json.Unmarshal(ingressInfo, &ingress)
	if err != nil {
		log.Println(err)
		return
	}

	for _, item := range ingress.Items {
		ipStr := getIpStrFromItem(item)
		for _, rule := range item.Spec.Rules {
			if rule.Host == "" {
				return
			}
			ingressInfo := IngressInfo{
				Endpoint:  rule.Host,
				Ip:        ipStr,
				Namespace: namespace,
			}
			clusterInfo.Ingresses = append(clusterInfo.Ingresses, ingressInfo)
		}
	}
}
func getIpStrFromItem(item Item) string {
	ipStr := ""
	if len(item.Status.LoadBalancer.Ingress) > 0 {
		for i, ingress := range item.Status.LoadBalancer.Ingress {
			ipStr += ingress.Ip
			if len(item.Status.LoadBalancer.Ingress) > i+1 {
				ipStr += ", "
			}
		}
	}
	return ipStr
}

func connectk8s() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	config.QPS = 50
	config.Burst = 150
	if err != nil {
		log.Println(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return clientset
}
