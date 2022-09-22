package cluster

type PodInfo struct {
	Namespace string
	Name      string
	Image     string
	Status    string
}
type IngressInfo struct {
	Endpoint  string
	Ip        string
	Namespace string
}
type DeploymentInfo struct {
	Namespace     string
	Name          string
	Replicas      int32
	ReadyReplicas int32
}
type NodeInfo map[string][]PodInfo
type ClusterInfo struct {
	Nodes       NodeInfo
	Ingresses   []IngressInfo
	Deployments []DeploymentInfo
	Timestamp   string
}
type Ingress struct {
	Items []Item
}
type Item struct {
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
