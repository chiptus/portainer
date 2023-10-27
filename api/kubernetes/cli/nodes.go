package cli

import (
	"context"

	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNodes get a list of all node names and ip addresses in the current k8s environment connection
func (kcl *KubeClient) GetNodes() ([]models.K8sNodes, error) {
	nodeList := make([]models.K8sNodes, 0)

	nodes, err := kcl.cli.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				nodeList = append(nodeList, models.K8sNodes{
					Name:    node.Name,
					Address: addr.Address,
				})
			}
		}
	}

	return nodeList, nil
}
