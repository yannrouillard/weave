package main

import (
	"fmt"
	"log"
	"strings"

	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

func getKubePeers() ([]string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	nodeList, err := c.Nodes().List(api.ListOptions{})
	if err != nil {
		// Fallback for cases (e.g. from kube-up.sh) where kube-proxy is not running on master
		config.Host = "http://localhost:8080"
		log.Print("error contacting APIServer: ", err, "; trying with fallback: ", config.Host)
		c, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
		nodeList, err = c.Nodes().List(api.ListOptions{})
	}

	if err != nil {
		return nil, err
	}
	addresses := make([]string, 0, len(nodeList.Items))
	for _, peer := range nodeList.Items {
		for _, addr := range peer.Status.Addresses {
			if addr.Type == "InternalIP" && addressLooksOK(addr.Address) {
				addresses = append(addresses, addr.Address)
			}
		}
	}
	return addresses, nil
}

// There isn't much documentation for what can appear in the 'address' field, but we have seen
// some installations with IPv6 addresses, which Weave Net cannot handle. So filter them out.
func addressLooksOK(address string) bool {
	if strings.Contains(address, ":") {
		return false
	}
	return true
}

func main() {
	peers, err := getKubePeers()
	if err != nil {
		log.Fatalf("Could not get peers: %v", err)
	}
	for _, addr := range peers {
		fmt.Println(addr)
	}
}
