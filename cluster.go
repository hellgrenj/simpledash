package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PodInfo struct {
	Namespace string
	Name      string
	Image     string
	Status    string
}
type IngressInfo struct {
	Endpoint string
	Ip       string
}
type NodeInfo map[string][]PodInfo
type ClusterInfo struct {
	Nodes     NodeInfo
	Ingresses []IngressInfo
}

func MonitorCluster(clusterInfoChan chan<- ClusterInfo) {
	clientset := connectk8s()
	for {
		clusterInfo := ClusterInfo{
			Nodes: make(NodeInfo),
		}
		sc := getContext()
		for _, namespace := range sc.Namespaces {
			getPodsByNamespace(clientset, &clusterInfo, namespace)
			getIngressInfo(clientset, &clusterInfo, namespace)
		}
		clusterInfoChan <- clusterInfo
		time.Sleep(time.Second * 10)
	}
}
func getPodsByNamespace(clientset *kubernetes.Clientset, clusterInfo *ClusterInfo, namespace string) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}
	for p := range pods.Items {

		podInfo := PodInfo{
			Namespace: pods.Items[p].Namespace,
			Name:      pods.Items[p].Name,
			Image:     pods.Items[p].Spec.Containers[0].Image,
			Status:    string(pods.Items[p].Status.Phase),
		}
		clusterInfo.Nodes[pods.Items[p].Spec.NodeName] = append(clusterInfo.Nodes[pods.Items[p].Spec.NodeName], podInfo)
	}
}

type Ingress struct {
	Items []struct {
		Spec struct {
			Rules []struct {
				Host string
			}
		}
		Status struct {
			LoadBalancer struct {
				Ingress []struct {
					Ip string
				}
			}
		}
	}
}

func getIngressInfo(clientset *kubernetes.Clientset, clusterInfo *ClusterInfo, namespace string) {
	ingresses := clientset.CoreV1().RESTClient().Get().AbsPath(fmt.Sprintf("/apis/networking.k8s.io/v1/namespaces/%s/ingresses", namespace)).Do(context.TODO())
	ingressInfo, err := ingresses.Raw()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(ingressInfo))
	var ingress Ingress
	json.Unmarshal(ingressInfo, &ingress)
	if err != nil {
		log.Println(err)
		return
	}

	for _, item := range ingress.Items {
		ip := item.Status.LoadBalancer.Ingress[0].Ip
		for _, rule := range item.Spec.Rules {
			if rule.Host == "" {
				return
			}
			ingressInfo := IngressInfo{
				Endpoint: rule.Host,
				Ip:       ip,
			}
			clusterInfo.Ingresses = append(clusterInfo.Ingresses, ingressInfo)
		}
	}
}

func connectk8s() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return clientset
}
