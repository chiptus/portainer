package microk8s

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
)

type (
	AddonPair struct {
		portaineree.Pair
		VersionAvailableFrom string `json:"versionAvailableFrom"`
		VersionAvailableTo   string `json:"versionAvailableTo"`
		Type                 string `json:"type"`
	}

	MicroK8sInfo struct {
		KubernetesVersions []portaineree.Pair `json:"kubernetesVersions"`
		AvailableAddons    []AddonPair        `json:"availableAddons"`
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
			Name                         string `json:"name"`
			Repository                   string `json:"repository"`
			Description                  string `json:"description"`
			MicroK8sVersionAvailableFrom string `json:"version"`
			Status                       string `json:"status"`
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
		Addons            []string
		KubernetesVersion string `json:"kubernetesVersion"`
		Scale             bool
	}
)

var MicroK8sVersions = []portaineree.Pair{
	{
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
	Name         string `json:"name"`
	Description  string `json:"description"`  // Notes/Extra info
	Tooltip      string `json:"tooltip"`      // used as a tooltip on the Addon page
	Arguments    string `json:"arguments"`    // used to pass arguments to the addon
	ArgumentType string `json:"argumentType"` // string
	Placeholder  string `json:"placeholder"`  // used as a placeholder for the argument input

	IsDefault   bool `json:"-"`
	IsAvailable bool `json:"-"`

	Type                         string `json:"type"`                         // if core/community
	MicroK8sVersionAvailableFrom string `json:"microK8sVersionAvailableFrom"` // microk8s version if the addon is available from a specific version
	MicroK8sVersionAvailableTo   string `json:"microK8sVersionAvailableTo"`   // microk8s version if the addon is unavailable from a specific version

	RequiredOnAllMasterNodes []string `json:"-"` // affected microk8s versions if the addon is required on all master nodes
}

type Addons []Addon

var AllAddons = Addons{
	{
		Name:                         "metrics-server",
		MicroK8sVersionAvailableFrom: "1.12",
		Tooltip:                      "metrics-server - adds the <a href='https://github.com/kubernetes-sigs/metrics-server' target='_blank'>Kubernetes Metrics Server</a> for API access to service metrics.",

		IsAvailable: true,
		Type:        "core",
	},
	{
		Name:                         "ingress",
		MicroK8sVersionAvailableFrom: "1.12",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-ingress' target='_blank'>ingress</a> - a simple ingress controller for external access.",

		IsAvailable: true,
		Type:        "core",
	},
	{
		Name:                         "host-access",
		MicroK8sVersionAvailableFrom: "1.19",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-host-access' target='_blank'>host-access</a> - provides a fixed IP for access to the host’s services.",

		Type:        "core",
		IsAvailable: true,
	},
	{
		Name:                         "cert-manager",
		MicroK8sVersionAvailableFrom: "1.25",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-cert-manager' target='_blank'>cert-manager</a> - certificate management for Kubernetes clusters.",

		IsAvailable: true,
		Type:        "core",
	},
	{
		Name:                         "gpu",
		MicroK8sVersionAvailableFrom: "1.12",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-gpu' target='_blank'>gpu</a> - enables support for GPU-accelerated workloads using the NVIDIA runtime.",

		IsAvailable: true,
		Type:        "core",
	},
	{
		Name:                         "hostpath-storage",
		MicroK8sVersionAvailableFrom: "1.12",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-hostpath-storage' target='_blank'>hostpath-storage</a> - creates a default storage class which allocates storage from a host directory. Note!: The add-on uses simple filesystem storage local to the node where it was added. Not suitable for a production environment or multi-node clusters.",

		IsAvailable: true,
		Type:        "core",
	},
	{
		Name:                         "observability",
		MicroK8sVersionAvailableFrom: "1.25",
		Tooltip:                      "observability - deploys the <a href='https://prometheus.io/docs/' target='_blank'>Kubernetes Prometheus Observability Stack.</a>",

		IsAvailable: true,
		Type:        "core",
	},
	{
		Name:                         "prometheus",
		MicroK8sVersionAvailableFrom: "1.14",
		MicroK8sVersionAvailableTo:   "1.24",
		Tooltip:                      "prometheus - deploys the <a href='https://prometheus.io/docs/' target='_blank'>Kubernetes Prometheus Operator</a>.",

		IsAvailable: true,
		Type:        "core",
	},
	{
		Name:                         "registry",
		MicroK8sVersionAvailableFrom: "1.12",
		Tooltip:                      "registry - deploys a private image registry and exposes it on localhost:32000.",

		IsAvailable: true,
		Type:        "core",
	},

	{
		Name:                         "dashboard",
		MicroK8sVersionAvailableFrom: "1.12",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-dashboard' target='_blank'>dashboard</a> - the standard Kubernetes Dashboard.",

		IsAvailable: true,
		Type:        "core",
	},

	{
		Name:                         "argocd",
		MicroK8sVersionAvailableFrom: "1.14",
		Tooltip:                      "argocd - deploys <a href='https://argo-cd.readthedocs.io/en/stable/' target='_blank'>Argo CD</a>, the declarative, GitOps continuous delivery tool for Kubernetes.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "cilium",
		MicroK8sVersionAvailableFrom: "1.15",
		Tooltip:                      "cilium - deploys <a href='http://docs.cilium.io/en/stable/intro/' target='_blank'>Cilium</a> to support Kubernetes network policies using eBPF.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "easyhaproxy",
		MicroK8sVersionAvailableFrom: "1.27",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-easyhaproxy' target='_blank'>easyhaproxy</a> - adds EasyHAProxy for automatic ingress.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "fluentd",
		MicroK8sVersionAvailableFrom: "1.13",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-fluentd' target='_blank'>fluentd<a> - deploy the <a href='https://www.elastic.co/guide/en/kibana/current/discover.html' target='_blank'>Elasticsearch-Fluentd-Kibana</a> logging and monitoring solution.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "gopaddle-lite",
		MicroK8sVersionAvailableFrom: "1.26",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-gopaddle-lite' target='_blank'>fluentd<a> - deploy the <a href='https://www.elastic.co/guide/en/kibana/current/discover.html' target='_blank'>Elasticsearch-Fluentd-Kibana</a> logging and monitoring solution.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "inaccel",
		MicroK8sVersionAvailableFrom: "1.24",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-inaccel' target='_blank'>inaccel</a> - FPGA and application lifecycle management.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "istio",
		MicroK8sVersionAvailableFrom: "1.12",
		Tooltip:                      "istio - adds the core <a href='https://istio.io/latest/docs/setup/platform-setup/microk8s/' target='_blank'>Istio</a> services (not available on arm64 arch).",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "jaeger",
		MicroK8sVersionAvailableFrom: "1.13",
		Tooltip:                      "jaeger - deploys the <a href='https://github.com/jaegertracing/jaeger-operator' target='_blank'>Jaeger Operator</a> - distributed tracing system for monitoring and troubleshooting microservices-based distributed systems. The simplest configuration is deployed - see <a href='https://www.jaegertracing.io/docs/' target='_blank'>documentation</a>.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "kata",
		MicroK8sVersionAvailableFrom: "1.22",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-kata' target='_blank'>kata</a> - adds <a href='https://katacontainers.io/' target='_blank'>Kata Containers</a> support - a secure container runtime with lightweight virtual machines.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "keda",
		MicroK8sVersionAvailableFrom: "1.20",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-keda' target='_blank'>keda</a> - deploys <a href='https://keda.sh/' target='_blank'>KEDA</a> - Kubernetes Event-driven Autoscaler.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "knative",
		MicroK8sVersionAvailableFrom: "1.15",
		Tooltip:                      "knative - adds the <a href='https://knative.dev/' target='_blank'>Knative</a> middleware to your cluster.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "kwasm",
		MicroK8sVersionAvailableFrom: "1.26",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-kwasm' target='_blank'>kwasm</a> - adds <a href='https://kwasm.sh/?dist=microk8s#Quickstart.' target='_blank'>Kwasm</a> - for WebAssembly support on your Kubernetes nodes.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "linkerd",
		MicroK8sVersionAvailableFrom: "1.19",
		Tooltip:                      "linkerd - deploys the linkerd service mesh.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "multus",
		MicroK8sVersionAvailableFrom: "1.19",
		Tooltip:                      "multus - adds Multus for multiple network capability.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "ondat",
		MicroK8sVersionAvailableFrom: "1.26",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-ondat' target='_blank'>ondat</a> - deploys <a href='https://docs.ondat.io/docs/' target='_blank'>Ondat</a> - a Kubernetes-native persistent storage platform.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "openfaas",
		MicroK8sVersionAvailableFrom: "1.21",
		Tooltip:                      "openfaas - deploys the OpenFaaS serverless functions framework. See configuration documentation.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "osm-edge",
		MicroK8sVersionAvailableFrom: "1.25",
		Tooltip:                      "osm-edge - Open Service Mesh Edge (OSM-Edge) fork from Open Service Mesh is a lightweight, extensible, Cloud Native service mesh built for Edge computing.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "shifu",
		MicroK8sVersionAvailableFrom: "1.27",
		Tooltip:                      "shifu - Kubernetes native IoT development framework.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "sosivio",
		MicroK8sVersionAvailableFrom: "1.26",
		Tooltip:                      "sosivio - deploys Sosivio predictive troubleshooting for Kubernetes.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "traefik",
		MicroK8sVersionAvailableFrom: "1.20",
		Tooltip:                      "traefik - adds the Traefik Kubernetes Ingress controller.",

		IsAvailable: true,
		Type:        "community",
	},

	{
		Name:                         "trivy",
		MicroK8sVersionAvailableFrom: "1.26",
		Tooltip:                      "trivy - deploys the Trivy open source security scanner for Kubernetes.",

		IsAvailable: true,
		Type:        "community",
	},

	// Default Addons
	{
		Name:                         "dns",
		MicroK8sVersionAvailableFrom: "1.12",
		Tooltip:                      "<a href='https://microk8s.io/docs/addon-dns' target='_blank'>dns</a> - deploys CoreDNS. It is recommended that this addon is always enabled.",

		IsAvailable: true,
		IsDefault:   true,
		Type:        "core",
	},
	{
		Name:                         "ha-cluster",
		MicroK8sVersionAvailableFrom: "1.19",
		Tooltip:                      "ha-cluster - allows for high availability on clusters with at least three nodes.",

		IsAvailable: true,
		IsDefault:   true,
		Type:        "core",
	},
	{
		Name:                         "helm",
		MicroK8sVersionAvailableFrom: "1.15",
		Tooltip:                      "helm - installs the <a href='https://helm.sh/' target='_blank'>Helm 3</a> package manager for Kubernetes",

		IsAvailable: true,
		IsDefault:   true,
		Type:        "core",
	},
	{
		Name:                         "helm3",
		MicroK8sVersionAvailableFrom: "1.18",
		Tooltip:                      "helm3 - transition addon introducing the <a href='https://helm.sh/' target='_blank'>Helm 3</a> package manager for Kubernetes.",

		IsAvailable: true,
		IsDefault:   true,
		Type:        "core",
	},
	{
		Name:                         "rbac",
		MicroK8sVersionAvailableFrom: "1.14",
		Tooltip:                      "rbac - enables Role Based Access Control for authorisation. Note that this is incompatible with some other add-ons.",

		IsAvailable: true,
		IsDefault:   true,
		Type:        "core",
	},
	{
		Name:                         "community",
		MicroK8sVersionAvailableFrom: "1.14",

		IsAvailable: true,
		IsDefault:   true,

		Type:                     "core",
		RequiredOnAllMasterNodes: []string{"1.27/stable"},
	},
}

func GetAllDefaultAddons() []string {
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

func (a Addons) GetNames() []string {
	var names []string
	for _, addon := range a {
		names = append(names, addon.Name)
	}
	return names
}
