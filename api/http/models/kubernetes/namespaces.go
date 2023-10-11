package models

import (
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/api/resource"
)

type K8sNamespaceDetails struct {
	Name              string                     `json:"Name"`
	Annotations       map[string]string          `json:"Annotations"`
	ResourceQuota     *K8sResourceQuota          `json:"ResourceQuota"`
	LoadBalancerQuota *K8sLoadBalancerQuota      `json:"LoadBalancerQuota"`
	StorageQuotas     map[string]K8sStorageQuota `json:"StorageQuotas"`
}

type K8sResourceQuota struct {
	Enabled bool   `json:"enabled"`
	Memory  string `json:"memory"`
	CPU     string `json:"cpu"`
}

type K8sLoadBalancerQuota struct {
	Enabled bool  `json:"enabled"`
	Limit   int64 `json:"limit"`
}

type K8sStorageQuota struct {
	Enabled bool   `json:"enabled"`
	Limit   string `json:"limit"`
}

func (r *K8sNamespaceDetails) Validate(request *http.Request) error {
	if r.ResourceQuota != nil && r.ResourceQuota.Enabled {
		_, err := resource.ParseQuantity(r.ResourceQuota.Memory)
		if err != nil {
			return fmt.Errorf("error parsing memory quota value: %w", err)
		}

		_, err = resource.ParseQuantity(r.ResourceQuota.CPU)
		if err != nil {
			return fmt.Errorf("error parsing cpu quota value: %w", err)
		}

	}

	if r.LoadBalancerQuota != nil && r.LoadBalancerQuota.Enabled {
		// must be a non negative integer
		if r.LoadBalancerQuota.Limit < 0 {
			return fmt.Errorf("load balancer quota limit must be a non negative integer")
		}
	}

	for _, storageClass := range r.StorageQuotas {
		if storageClass.Enabled {
			_, err := resource.ParseQuantity(storageClass.Limit)
			if err != nil {
				return fmt.Errorf("error parsing storage class quota limit: %w", err)
			}
		}
	}

	return nil
}
