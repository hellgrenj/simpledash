package cluster

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
