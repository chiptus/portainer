package util

import (
	"strings"

	kubeModels "github.com/portainer/portainer-ee/api/http/models/kubernetes"
)

func NodeListToIpList(nodes []kubeModels.K8sNodes) []string {
	flat := []string{}
	for _, node := range nodes {
		flat = append(flat, node.Address)
	}
	return flat
}

func UrlToMasterNode(url string) string {
	u := strings.Split(url, ":")
	return u[0]
}
