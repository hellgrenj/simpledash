package cluster

import (
	"testing"
)

func TestGetIpsFromItem(t *testing.T) {
	item := Item{}
	item.Status.LoadBalancer.Ingress = []struct{ Ip string }{{Ip: "172.123.1.173"}, {Ip: "172.123.1.174"}}
	ipStr := getIpsFromItem(item)
	expectedResult := "172.123.1.173, 172.123.1.174"
	if ipStr != expectedResult {
		t.Errorf("Expected %s got %s", expectedResult, ipStr)
	}
	oneIp := "172.23.1.175"
	item.Status.LoadBalancer.Ingress = []struct{ Ip string }{{Ip: oneIp}}
	ipStr = getIpsFromItem(item)
	if oneIp != ipStr {
		t.Errorf("Expected %s got %s", oneIp, ipStr)
	}
}
