package cloud

import (
	"reflect"
	"testing"
)

func TestParseAddonResponse(t *testing.T) {
	type test struct {
		input string
		want  []string
	}

	tests := []test{
		{
			// Empty test.
		},
		{
			input: `high-availability: no
	  datastore master nodes: 127.0.0.1:19001
	  datastore standby nodes: none
	addons:
	  enabled:
		dns                  # (core) CoreDNS
		ha-cluster           # (core) Configure high availability on the current node
		helm                 # (core) Helm - the package manager for Kubernetes
		helm3                # (core) Helm 3 - the package manager for Kubernetes
		hostpath-storage     # (core) Storage class; allocates storage from host directory
		metrics-server       # (core) K8s Metrics Server for API access to service metrics
		rbac                 # (core) Role-Based Access Control for authorisation
		storage              # (core) Alias to hostpath-storage add-on, deprecated
	  disabled:
		cert-manager         # (core) Cloud native certificate management
		community            # (core) The community addons repository
		dashboard            # (core) The Kubernetes dashboard
		gpu                  # (core) Automatic enablement of Nvidia CUDA
		host-access          # (core) Allow Pods connecting to Host services smoothly
		ingress              # (core) Ingress controller for external access
		kube-ovn             # (core) An advanced network fabric for Kubernetes
		mayastor             # (core) OpenEBS MayaStor
		metallb              # (core) Loadbalancer for your Kubernetes cluster
		observability        # (core) A lightweight observability stack for logs, traces and metrics
		prometheus           # (core) Prometheus operator for monitoring and logging
		registry             # (core) Private image registry exposed on localhost:32000`,

			want: []string{
				"dns",
				"ha-cluster",
				"helm",
				"helm3",
				"hostpath-storage",
				"metrics-server",
				"rbac",
				"storage",
			},
		},
	}

	for _, tc := range tests {
		got := parseAddonResponse(tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Errorf("want: %v\ngot: %v", tc.want, got)
		}
	}
}
