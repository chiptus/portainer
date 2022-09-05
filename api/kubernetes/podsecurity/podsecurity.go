package podsecurity

import "time"

type (
	PodSecurityRuleID int
	PodSecurityRule   struct {
		ID                       PodSecurityRuleID                   `json:"id"`
		Enabled                  bool                                `json:"enabled"`
		EndpointID               int                                 `json:"endPointID"`
		PrivilegedContainers     PodSecurityPrivilegedContainers     `json:"privilegedContainers"`
		HostNamespaces           PodSecurityHostNamespaces           `json:"hostNamespaces"`
		HostNetworkingPorts      PodSecurityHostNetworkingPorts      `json:"hostPorts"`
		VolumeTypes              PodSecurityVolumeTypes              `json:"volumeTypes"`
		HostFilesystem           PodSecurityHostFilesystem           `json:"hostFilesystem"`
		AllowFlexVolumes         PodSecurityAllowFlexVolumes         `json:"allowFlexVolumes"`
		Users                    PodSecurityUsers                    `json:"users"`
		AllowPrivilegeEscalation PodSecurityAllowPrivilegeEscalation `json:"allowPrivilegeEscalation"`
		Capabilities             PodSecurityCapabilities             `json:"capabilities"`
		Selinux                  PodSecuritySelinux                  `json:"selinux"`
		AllowProcMount           PodSecurityAllowProcMount           `json:"allowProcMount"`
		AppArmour                PodSecurityAppArmour                `json:"appArmor"`
		ReadOnlyRootFileSystem   PodSecurityReadOnlyRootFileSystem   `json:"readOnlyRootFileSystem"`
		SecComp                  PodSecuritySecComp                  `json:"secComp"`
		ForbiddenSysctlsList     PodSecurityForbiddenSysctlsList     `json:"forbiddenSysctlsList"`
	}
	PodSecurityAllowPrivilegeEscalation struct {
		Enabled bool `json:"enabled"`
	}
	PodSecurityReadOnlyRootFileSystem struct {
		Enabled bool `json:"enabled"`
	}
	PodSecurityPrivilegedContainers struct {
		Enabled bool `json:"enabled"`
	}
	PodSecurityHostNamespaces struct {
		Enabled bool `json:"enabled"`
	}

	PodSecurityHostNetworkingPorts struct {
		Enabled     bool `json:"enabled"`
		HostNetwork bool `json:"hostNetwork"`
		Max         int  `json:"max"`
		Min         int  `json:"min"`
	}
	PodSecurityVolumeTypes struct {
		Enabled      bool     `json:"enabled"`
		AllowedTypes []FSType `json:"allowedTypes"`
	}

	FSType string

	PodSecurityAllowedPaths struct {
		PathPrefix string `json:"pathPrefix"`
		Readonly   bool   `json:"readonly"`
	}
	PodSecurityHostFilesystem struct {
		Enabled      bool                      `json:"enabled"`
		AllowedPaths []PodSecurityAllowedPaths `json:"allowedPaths"`
	}
	PodSecurityAllowFlexVolumes struct {
		Enabled        bool     `json:"enabled"`
		AllowedVolumes []string `json:"allowedVolumes"`
	}
	PodSecurityRunAsUser struct {
		Type    RunAsUserStrategy    `json:"type"`
		Idrange []PodSecurityIdrange `json:"idrange"`
	}
	PodSecurityIdrange struct {
		Max int `json:"max"`
		Min int `json:"min"`
	}
	RunAsUserStrategy string

	PodSecurityRunAsGroup struct {
		Type    RunAsGroupStrategy   `json:"type"`
		Idrange []PodSecurityIdrange `json:"idrange"`
	}
	RunAsGroupStrategy string

	PodSecuritySupplementalGroups struct {
		Type    SupplementalGroupsStrategyType `json:"type"`
		Idrange []PodSecurityIdrange           `json:"idrange"`
	}
	SupplementalGroupsStrategyType string

	PodSecurityFsGroups struct {
		Type    FSGroupStrategyType  `json:"type"`
		Idrange []PodSecurityIdrange `json:"idrange"`
	}
	FSGroupStrategyType string

	PodSecurityUsers struct {
		Enabled            bool                          `json:"enabled"`
		RunAsUser          PodSecurityRunAsUser          `json:"runAsUser"`
		RunAsGroup         PodSecurityRunAsGroup         `json:"runAsGroup"`
		SupplementalGroups PodSecuritySupplementalGroups `json:"supplementalGroups"`
		FsGroups           PodSecurityFsGroups           `json:"fsGroups"`
	}
	PodSecurityCapabilities struct {
		Enabled                  bool     `json:"enabled"`
		AllowedCapabilities      []string `json:"allowedCapabilities"`
		RequiredDropCapabilities []string `json:"requiredDropCapabilities"`
	}
	PodSecurityAllowedCapabilities struct {
		Level string `json:"level"`
		Role  string `json:"role"`
		Type  string `json:"type"`
		User  string `json:"user"`
	}
	PodSecuritySelinux struct {
		Enabled             bool                             `json:"enabled"`
		AllowedCapabilities []PodSecurityAllowedCapabilities `json:"allowedCapabilities"`
	}
	PodSecurityAllowProcMount struct {
		Enabled       bool   `json:"enabled"`
		ProcMountType string `json:"procMountType"`
	}
	PodSecurityAppArmour struct {
		Enabled       bool     `json:"enabled"`
		AppArmourType []string `json:"AppArmorType"`
	}
	PodSecuritySecComp struct {
		Enabled     bool     `json:"enabled"`
		SecCompType []string `json:"secCompType"`
	}
	PodSecurityForbiddenSysctlsList struct {
		Enabled                  bool     `json:"enabled"`
		RequiredDropCapabilities []string `json:"requiredDropCapabilities"`
	}
	PodSecurityConstraintCommon struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
		Spec struct {
			Match struct {
				Kinds []struct {
					APIGroups []string `yaml:"apiGroups"`
					Kinds     []string `yaml:"kinds"`
				} `yaml:"kinds"`
				ExcludedNamespaces []string `yaml:"excludedNamespaces,omitempty"`
			} `yaml:"match"`
			Parameters interface{} `yaml:"parameters,omitempty"`
		} `yaml:"spec"`
	}
)

const (
	AzureFile             FSType = "azureFile"
	Flocker               FSType = "flocker"
	FlexVolume            FSType = "flexVolume"
	HostPath              FSType = "hostPath"
	EmptyDir              FSType = "emptyDir"
	GCEPersistentDisk     FSType = "gcePersistentDisk"
	AWSElasticBlockStore  FSType = "awsElasticBlockStore"
	GitRepo               FSType = "gitRepo"
	Secret                FSType = "secret"
	NFS                   FSType = "nfs"
	ISCSI                 FSType = "iscsi"
	Glusterfs             FSType = "glusterfs"
	PersistentVolumeClaim FSType = "persistentVolumeClaim"
	RBD                   FSType = "rbd"
	Cinder                FSType = "cinder"
	CephFS                FSType = "cephFS"
	DownwardAPI           FSType = "downwardAPI"
	FC                    FSType = "fc"
	ConfigMap             FSType = "configMap"
	VsphereVolume         FSType = "vsphereVolume"
	Quobyte               FSType = "quobyte"
	AzureDisk             FSType = "azureDisk"
	PhotonPersistentDisk  FSType = "photonPersistentDisk"
	StorageOS             FSType = "storageos"
	Projected             FSType = "projected"
	PortworxVolume        FSType = "portworxVolume"
	ScaleIO               FSType = "scaleIO"
	CSI                   FSType = "csi"
	Ephemeral             FSType = "ephemeral"
	All                   FSType = "*"
)
const (
	// RunAsUserStrategyMustRunAs means that container must run as a particular uid.
	RunAsUserStrategyMustRunAs RunAsUserStrategy = "MustRunAs"
	// RunAsUserStrategyMustRunAsNonRoot means that container must run as a non-root uid.
	RunAsUserStrategyMustRunAsNonRoot RunAsUserStrategy = "MustRunAsNonRoot"
	// RunAsUserStrategyRunAsAny means that container may make requests for any uid.
	RunAsUserStrategyRunAsAny RunAsUserStrategy = "RunAsAny"
)
const (
	// FSGroupStrategyMayRunAs means that container does not need to have FSGroup of X applied.
	// However, when FSGroups are specified, they have to fall in the defined range.
	FSGroupStrategyMayRunAs FSGroupStrategyType = "MayRunAs"
	// FSGroupStrategyMustRunAs meant that container must have FSGroup of X applied.
	FSGroupStrategyMustRunAs FSGroupStrategyType = "MustRunAs"
	// FSGroupStrategyRunAsAny means that container may make requests for any FSGroup labels.
	FSGroupStrategyRunAsAny FSGroupStrategyType = "RunAsAny"
)
const (
	// RunAsGroupStrategyMayRunAs means that container does not need to run with a particular gid.
	// However, when RunAsGroup are specified, they have to fall in the defined range.
	RunAsGroupStrategyMayRunAs RunAsGroupStrategy = "MayRunAs"
	// RunAsGroupStrategyMustRunAs means that container must run as a particular gid.
	RunAsGroupStrategyMustRunAs RunAsGroupStrategy = "MustRunAs"
	// RunAsUserStrategyRunAsAny means that container may make requests for any gid.
	RunAsGroupStrategyRunAsAny RunAsGroupStrategy = "RunAsAny"
)
const (
	// SupplementalGroupsStrategyMayRunAs means that container does not need to run with a particular gid.
	// However, when gids are specified, they have to fall in the defined range.
	SupplementalGroupsStrategyMayRunAs SupplementalGroupsStrategyType = "MayRunAs"
	// SupplementalGroupsStrategyMustRunAs means that container must run as a particular gid.
	SupplementalGroupsStrategyMustRunAs SupplementalGroupsStrategyType = "MustRunAs"
	// SupplementalGroupsStrategyRunAsAny means that container may make requests for any gid.
	SupplementalGroupsStrategyRunAsAny SupplementalGroupsStrategyType = "RunAsAny"
)

var PodSecurityConstraintsMap = map[string]string{
	"K8sPSPAllowPrivilegeEscalationContainer": "allow-privilege-escalation",
	"K8sPSPAppArmor":               "apparmor",
	"K8sPSPCapabilities":           "capabilities",
	"K8sPSPFlexVolumes":            "flexvolume-drivers",
	"K8sPSPForbiddenSysctls":       "forbidden-sysctls",
	"K8sPSPFSGroup":                "fsgroup",
	"K8sPSPHostFilesystem":         "host-filesystem",
	"K8sPSPHostNamespace":          "host-namespaces",
	"K8sPSPHostNetworkingPorts":    "host-network-ports",
	"K8sPSPPrivilegedContainer":    "privileged-containers",
	"K8sPSPProcMount":              "proc-mount",
	"K8sPSPReadOnlyRootFilesystem": "read-only-root-filesystem",
	"K8sPSPSeccomp":                "seccomp",
	"K8sPSPSELinuxV2":              "selinux",
	"K8sPSPAllowedUsers":           "users",
	"K8sPSPVolumeTypes":            "volumes"}

var (
	GateKeeperFile                   = "gatekeeper.yaml"
	GateKeeperExcludedNamespacesFile = "excluded-namespaces.yaml"
	GateKeeperNameSpace              = "gatekeeper-system"
	GateKeeperPodName                = "gatekeeper-controller-manager"
	GateKeeperSelector               = "gatekeeper.sh/operation"
	GateKeeperInterval               = 2 * time.Second
	GateKeeperTimeOut                = 300 * time.Second
)
