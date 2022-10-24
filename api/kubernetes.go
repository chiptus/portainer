package portaineree

func KubernetesDefault() KubernetesData {
	return KubernetesData{
		Configuration: KubernetesConfiguration{
			UseLoadBalancer:                 false,
			UseServerMetrics:                false,
			EnableResourceOverCommit:        true,
			ResourceOverCommitPercentage:    20,
			StorageClasses:                  []KubernetesStorageClassConfig{},
			IngressClasses:                  []KubernetesIngressClassConfig{},
			RestrictDefaultNamespace:        false,
			IngressAvailabilityPerNamespace: false,
			AllowNoneIngressClass:           false,
		},
		Snapshots: []KubernetesSnapshot{},
	}
}
