package cli

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetCPUCount get a list of all node names and ip addresses in the current k8s
// environment connection
func (kcl *KubeClient) GetCPUCount() (int, error) {
	var cpuCount int

	nodes, err := kcl.cli.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return cpuCount, err
	}

	for _, node := range nodes.Items {
		quantity := node.Status.Capacity.Cpu()
		if quantity != nil {
			cpu, ok := quantity.AsInt64()
			if ok {
				cpuCount += int(cpu)
			}
		}
	}

	return cpuCount, nil
}
