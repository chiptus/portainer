package portaineree

import portainer "github.com/portainer/portainer/api"

func KubernetesDefault() KubernetesData {
	return KubernetesData{
		Configuration: KubernetesConfiguration{
			UseLoadBalancer:                 false,
			UseServerMetrics:                false,
			EnableResourceOverCommit:        true,
			ResourceOverCommitPercentage:    20,
			StorageClasses:                  []portainer.KubernetesStorageClassConfig{},
			IngressClasses:                  []portainer.KubernetesIngressClassConfig{},
			RestrictDefaultNamespace:        false,
			IngressAvailabilityPerNamespace: false,
			RestrictStandardUserIngressW:    false,
			AllowNoneIngressClass:           false,
		},
		Snapshots: []portainer.KubernetesSnapshot{},
	}
}
