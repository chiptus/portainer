package microk8s

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
)

type (
	MicroK8sInfo struct {
		KubernetesVersions []portaineree.Pair `json:"kubernetesVersions"`
		AvailableAddons    Addons             `json:"availableAddons"`
		RequiredAddons     []string           `json:"requiredAddons"`
	}

	MicroK8sInstalledAddons []string

	microk8sClusterJoinInfo struct {
		Token string   `json:"token"`
		URLS  []string `json:"urls"`
	}

	Microk8sStatusResponse struct {
		Microk8S struct {
			Running bool `json:"running"`
		} `json:"microk8s"`
		HighAvailability struct {
			Enabled bool `json:"enabled"`
			Nodes   []struct {
				Address string `json:"address"`
				Role    string `json:"role"`
			} `json:"nodes"`
		} `json:"highAvailability"`
		Addons []struct {
			Name        string `json:"name"`
			Repository  string `json:"repository"`
			Description string `json:"description"`
			Version     string `json:"version"` // addon version
			Status      string `json:"status"`
			Arguments   string `json:"arguments"` // Read from the endpoint CloudProvider
		} `json:"addons"`
		CurrentVersion     string             `json:"currentVersion"`
		KubernetesVersions []portaineree.Pair `json:"kubernetesVersions"`
		RequiredAddons     []string           `json:"requiredAddons"`
	}

	Microk8sNodeStatusResponse struct {
		Status string `json:"status"`
	}

	Microk8sProvisioningClusterRequest struct {
		EnvironmentID     portaineree.EndpointID `json:"environmentID"`
		Credentials       *models.CloudCredential
		MasterNodes       []string
		WorkerNodes       []string
		Addons            AddonsWithArgs
		KubernetesVersion string `json:"kubernetesVersion"`
		Scale             bool
	}
)

var MicroK8sVersions = []portaineree.Pair{
	{
		Name:  "1.28/stable",
		Value: "1.28/stable",
	}, {
		Name:  "1.27/stable",
		Value: "1.27/stable",
	},
	{
		Name:  "1.26/stable",
		Value: "1.26/stable",
	},
	{
		Name:  "1.25/stable",
		Value: "1.25/stable",
	},
	{
		Name:  "1.24/stable",
		Value: "1.24/stable",
	},
}

type Addon struct {
	Name    string `json:"label"`   // FE uses label as the addon name for dropdown
	Tooltip string `json:"tooltip"` // used as a tooltip on the Addon page

	ArgumentsType     string `json:"argumentsType"` // "": not required, optional: optional arguments, required: required arguments
	ArgumentSeparator string `json:"-"`             // used to separate arguments
	Placeholder       string `json:"placeholder"`   // used as a placeholder for the argument input

	IsDefault   bool `json:"isDefault"`
	IsAvailable bool `json:"-"`

	VersionAvailableTo   string `json:"versionAvailableTo"`   // microk8s version if the addon is unavailable from a specific version
	Repository           string `json:"repository"`           // if core/community
	VersionAvailableFrom string `json:"versionAvailableFrom"` // microk8s version if the addon is available from a specific version

	RequiredOn        string   `json:"-"` // Which nodes does this addon need installed on: "masters", "all", or a blank string for just the connected node.
	InstallCommands   []string `json:"-"`
	UninstallCommands []string `json:"-"`

	SkipUpgrade bool   `json:"skipUpgrade"`
	Info        string `json:"info"` // used to display additional information on the Addon page
}

type Addons []Addon

var AllAddons = Addons{
	{
		Name:                 "metrics-server",
		VersionAvailableFrom: "1.12",
		Tooltip:              "metrics-server - adds the <a href='https://github.com/kubernetes-sigs/metrics-server' target='_blank'>Kubernetes Metrics Server</a> for API access to service metrics.",

		IsAvailable: true,
		Repository:  "core",
	},
	{
		Name:                 "ingress",
		VersionAvailableFrom: "1.12",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-ingress' target='_blank'>ingress</a> - a simple ingress controller for external access.",

		IsAvailable: true,
		Repository:  "core",
	},
	{
		Name:                 "host-access",
		VersionAvailableFrom: "1.19",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-host-access' target='_blank'>host-access</a> - provides a fixed IP for access to the host’s services.",

		Repository:  "core",
		IsAvailable: true,

		RequiredOn:        "masters",
		ArgumentsType:     "optional",
		Placeholder:       "ip=10.0.1.2",
		ArgumentSeparator: ":",
	},
	{
		Name:                 "cert-manager",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-cert-manager' target='_blank'>cert-manager</a> - certificate management for Kubernetes clusters.",
		VersionAvailableFrom: "1.25",

		IsAvailable: true,
		Repository:  "core",
	},
	{
		Name:                 "gpu",
		VersionAvailableFrom: "1.12",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-gpu' target='_blank'>gpu</a> - enables support for GPU-accelerated workloads using the NVIDIA runtime.",

		ArgumentsType: "optional",
		Placeholder:   "--driver host",
		IsAvailable:   true,
		Repository:    "core",
	},
	{
		Name:                 "hostpath-storage",
		VersionAvailableFrom: "1.12",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-hostpath-storage' target='_blank'>hostpath-storage</a> - creates a default storage class which allocates storage from a host directory. Note!: The add-on uses simple filesystem storage local to the node where it was added. Not suitable for a production environment or multi-node clusters.",

		IsAvailable: true,
		Repository:  "core",

		SkipUpgrade: true,
	},
	{
		Name:                 "observability",
		VersionAvailableFrom: "1.25",
		Tooltip:              "observability - deploys the <a href='https://prometheus.io/docs/' target='_blank'>Kubernetes Prometheus Observability Stack.</a>",

		IsAvailable: true,
		Repository:  "core",
	},
	{
		Name:                 "prometheus",
		VersionAvailableFrom: "1.14",
		VersionAvailableTo:   "1.24",
		Tooltip:              "prometheus - deploys the <a href='https://prometheus.io/docs/' target='_blank'>Kubernetes Prometheus Operator</a>.",

		IsAvailable: true,
		Repository:  "core",
	},
	{
		Name:                 "registry",
		VersionAvailableFrom: "1.12",
		Tooltip:              "registry - deploys a private image registry and exposes it on localhost:32000.",

		IsAvailable: true,
		Repository:  "core",
	},
	{
		Name:                 "dashboard",
		VersionAvailableFrom: "1.12",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-dashboard' target='_blank'>dashboard</a> - the standard Kubernetes Dashboard.",

		IsAvailable: true,
		Repository:  "core",
	},
	{
		Name:                 "metallb",
		VersionAvailableFrom: "1.17",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-metallb' target='_blank'>metallb</a> - deploys the <a href='https://metallb.universe.tf/' target='_blank'>MetalLB load balancer</a>. Note that currently this does not work on macOS, due to network filtering.",

		ArgumentsType:     "required",
		ArgumentSeparator: ":",

		Placeholder: "10.64.140.43-10.64.140.49,10.64.141.53-10.64.141.59,10.12.13.0/24",
		IsAvailable: true,
		Repository:  "core",
	},
	{
		Name:                 "argocd",
		VersionAvailableFrom: "1.14",
		Tooltip:              "argocd - deploys <a href='https://argo-cd.readthedocs.io/en/stable/' target='_blank'>Argo CD</a>, the declarative, GitOps continuous delivery tool for Kubernetes.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "cilium",
		VersionAvailableFrom: "1.15",
		Tooltip:              "cilium - deploys <a href='http://docs.cilium.io/en/stable/intro/' target='_blank'>Cilium</a> to support Kubernetes network policies using eBPF.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "easyhaproxy",
		VersionAvailableFrom: "1.27",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-easyhaproxy' target='_blank'>easyhaproxy</a> - adds EasyHAProxy for automatic ingress.",

		IsAvailable: true,
		Repository:  "community",

		ArgumentsType: "optional",
		Placeholder:   "--nodeport",
	},
	{
		Name:                 "fluentd",
		VersionAvailableFrom: "1.13",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-fluentd' target='_blank'>fluentd</a> - deploy the <a href='https://www.elastic.co/guide/en/kibana/current/discover.html' target='_blank'>Elasticsearch-Fluentd-Kibana</a> logging and monitoring solution.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "gopaddle-lite",
		VersionAvailableFrom: "1.26",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-gopaddle-lite' target='_blank'>gopaddle-lite</a> - deploys the <a href='https://help.gopaddle.io/overview/getting-started/installing-community-edition/microk8s-addon/on-ubuntu' target='_blank'>gopaddle lite</a> no-code platform for Kubernetes developers.",

		IsAvailable: true,
		Repository:  "community",

		ArgumentsType: "optional",
		Placeholder:   "-i 130.198.9.42 -v 4.2.5",
	},
	{
		Name:                 "inaccel",
		VersionAvailableFrom: "1.24",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-inaccel' target='_blank'>inaccel</a> - FPGA and application lifecycle management.",

		IsAvailable: true,
		Repository:  "community",
	},

	{
		Name:                 "istio",
		VersionAvailableFrom: "1.12",
		Tooltip:              "istio - adds the core <a href='https://istio.io/latest/docs/setup/platform-setup/microk8s/' target='_blank'>Istio</a> services (not available on arm64 arch).",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "jaeger",
		VersionAvailableFrom: "1.13",
		Tooltip:              "jaeger - deploys the <a href='https://github.com/jaegertracing/jaeger-operator' target='_blank'>Jaeger Operator</a> - distributed tracing system for monitoring and troubleshooting microservices-based distributed systems. The simplest configuration is deployed - see <a href='https://www.jaegertracing.io/docs/' target='_blank'>documentation</a>.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "kata",
		VersionAvailableFrom: "1.22",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-kata' target='_blank'>kata</a> - adds <a href='https://katacontainers.io/' target='_blank'>Kata Containers</a> support - a secure container runtime with lightweight virtual machines.",

		IsAvailable: true,
		Repository:  "community",
		RequiredOn:  "masters",

		ArgumentsType: "optional",
		Placeholder:   "--runtime-path=/path/to/runtime",
	},
	{
		Name:                 "keda",
		VersionAvailableFrom: "1.20",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-keda' target='_blank'>keda</a> - deploys <a href='https://keda.sh/' target='_blank'>KEDA</a> - Kubernetes Event-driven Autoscaler.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "knative",
		VersionAvailableFrom: "1.15",
		Tooltip:              "knative - adds the <a href='https://knative.dev/' target='_blank'>Knative</a> middleware to your cluster.",

		IsAvailable: true,
		Repository:  "community",
	},

	{
		Name:                 "kwasm",
		VersionAvailableFrom: "1.26",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-kwasm' target='_blank'>kwasm</a> - adds <a href='https://kwasm.sh/?dist=microk8s#Quickstart.' target='_blank'>Kwasm</a> - for WebAssembly support on your Kubernetes nodes.",

		IsAvailable: true,
		Repository:  "community",
	},

	{
		Name:                 "linkerd",
		VersionAvailableFrom: "1.19",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-linkerd' target='_blank'>linkerd</a> - deploys the <a href='https://linkerd.io/2/overview/' target='_blank'>linkerd</a> service mesh.",

		IsAvailable: true,
		Repository:  "community",
	},

	{
		Name:                 "multus",
		VersionAvailableFrom: "1.19",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-multus' target='_blank'>multus</a> - adds <a href='https://github.com/k8snetworkplumbingwg/multus-cni' target='_blank'>Multus</a> for multiple network capability.",

		IsAvailable: true,
		Repository:  "community",
		RequiredOn:  "masters",
	},

	{
		Name:                 "ondat",
		VersionAvailableFrom: "1.26",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-ondat' target='_blank'>ondat</a> - deploys <a href='https://docs.ondat.io/docs/' target='_blank'>Ondat</a> - a Kubernetes-native persistent storage platform.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "openfaas",
		VersionAvailableFrom: "1.21",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-openfaas' target='_blank'>openfaas</a> - deploys the <a href='https://www.openfaas.com/' target='_blank'>OpenFaaS</a> serverless functions framework. See <a href='https://github.com/openfaas/faas-netes/tree/master/chart/openfaas#configuration' target='_blank'>configuration documentation</a>.",

		IsAvailable: true,
		Repository:  "community",

		ArgumentsType: "optional",
		Placeholder:   "--no-auth --operator -f values.yaml",
	},
	{
		Name:                 "osm-edge",
		VersionAvailableFrom: "1.25",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-osm' target='_blank'>osm-edge</a> - Open Service Mesh Edge (OSM-Edge) fork from Open Service Mesh is a lightweight, extensible, Cloud Native service mesh built for Edge computing.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "parking",
		VersionAvailableFrom: "1.27",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-parking' target='_blank'>parking</a> - Parking for static sites. A comma-separated list of domains must be specified.",

		IsAvailable: true,
		Repository:  "community",

		ArgumentsType: "required",
		Placeholder:   "<comma-separate-list-of-domains-to-be-parked>",
	},
	{
		Name:                 "shifu",
		VersionAvailableFrom: "1.27",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-shifu' target='_blank'>shifu</a> - Kubernetes native IoT development framework.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "sosivio",
		VersionAvailableFrom: "1.26",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-sosivio' target='_blank'>sosivio</a> - deploys <a href='https://docs.sosiv.io/' target='_blank'>Sosivio</a> predictive troubleshooting for Kubernetes.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "traefik",
		VersionAvailableFrom: "1.20",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-traefik' target='_blank'>traefik</a> - adds the <a href='https://doc.traefik.io/traefik/providers/kubernetes-ingress/' target='_blank'>Traefik Kubernetes Ingress controller</a>.",

		IsAvailable: true,
		Repository:  "community",
	},
	{
		Name:                 "trivy",
		VersionAvailableFrom: "1.26",
		Tooltip:              "<a href='https://discuss.kubernetes.io/t/addon-trivy/23797' target='_blank'>trivy</a> - deploys the <a href='https://aquasecurity.github.io/trivy/' target='_blank'>Trivy</a> open source security scanner for Kubernetes.",

		IsAvailable: true,
		Repository:  "community",
	},

	// Default Addons
	{
		Name:                 "dns",
		VersionAvailableFrom: "1.12",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-dns' target='_blank'>dns</a> - deploys CoreDNS. It is recommended that this addon is always enabled.",

		IsAvailable: true,
		IsDefault:   true,
		Repository:  "core",

		ArgumentsType:     "optional",
		ArgumentSeparator: ":",
	},
	{
		Name:                 "ha-cluster",
		VersionAvailableFrom: "1.19",
		Tooltip:              "ha-cluster - allows for high availability on clusters with at least three nodes.",

		IsAvailable: true,
		IsDefault:   true,
		Repository:  "core",
	},
	{
		Name:                 "helm",
		VersionAvailableFrom: "1.15",
		Tooltip:              "helm - installs the <a href='https://helm.sh/' target='_blank'>Helm 3</a> package manager for Kubernetes",

		IsAvailable: true,
		IsDefault:   true,
		Repository:  "core",
	},
	{
		Name:                 "helm3",
		VersionAvailableFrom: "1.18",
		Tooltip:              "helm3 - transition addon introducing the <a href='https://helm.sh/' target='_blank'>Helm 3</a> package manager for Kubernetes.",

		IsAvailable: true,
		IsDefault:   true,
		Repository:  "core",
	},
	{
		Name:                 "rbac",
		VersionAvailableFrom: "1.14",
		Tooltip:              "rbac - enables Role Based Access Control for authorisation. Note that this is incompatible with some other add-ons.",

		IsAvailable: true,
		IsDefault:   true,
		Repository:  "core",
	},
	{
		Name:                 "community",
		VersionAvailableFrom: "1.14",

		IsAvailable: true,
		IsDefault:   true,

		Repository: "core",
		RequiredOn: "masters",
	},
	{
		Name:                 "openebs",
		VersionAvailableFrom: "1.21",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-openebs' target='_blank'>OpenEBS</a> - the most widely deployed and easy to use open-source storage solution for Kubernetes.",

		IsAvailable: true,
		Repository:  "community",
		RequiredOn:  "all",
		InstallCommands: []string{
			"systemctl enable iscsid",
			"systemctl start iscsid",
		},
		UninstallCommands: []string{
			"systemctl disable iscsid",
			"systemctl stop iscsid",
		},
	},
	{
		Name:                 "mayastor",
		VersionAvailableFrom: "1.24",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-mayastor' target='_blank'>mayastor</a> - multi-node zero-ops storage option powered by <a href='https://github.com/openebs/mayastor' target='_blank'>Mayastor</a>.",

		IsAvailable: true,
		Repository:  "core",
		RequiredOn:  "all",

		ArgumentsType: "required",
		Placeholder:   "--default-pool-size 20G",

		Info: "To enable mayastor, ensure all nodes meet <a href='https://microk8s.io/docs/addon-mayastor' target='_blank'>requirements</a>, such as enabling HugePages with at least 1024 MB.",
	},
	{
		Name:                 "nfs",
		VersionAvailableFrom: "1.25",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-nfs' target='_blank'>nfs</a> - Enables NFS Ganesha Server and Provisioner to MicroK8s, running as a pod.",

		IsAvailable: true,
		Repository:  "community",
		RequiredOn:  "all",
		InstallCommands: []string{
			"apt install -y nfs-common",
		},
	},
	{
		Name:                 "minio",
		VersionAvailableFrom: "1.26",
		Tooltip:              "<a href='https://microk8s.io/docs/addon-minio' target='_blank'>minio</a> - Enables Minio High Performance Object Storage.",

		IsAvailable: true,
		Repository:  "community",
		RequiredOn:  "all",
	},
}

func GetDefaultAddons() []string {
	var addons []string
	for _, addon := range AllAddons {
		if addon.IsAvailable && addon.IsDefault {
			addons = append(addons, addon.Name)
		}
	}
	return addons
}

func GetAllAvailableAddons() Addons {
	var addons []Addon
	for _, addon := range AllAddons {
		if addon.IsAvailable && !addon.IsDefault {
			addons = append(addons, addon)
		}
	}
	return addons
}
