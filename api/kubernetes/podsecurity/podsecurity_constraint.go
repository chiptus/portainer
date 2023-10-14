package podsecurity

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type PodSecurityConstraint struct {
	userID              portainer.UserID
	endpoint            *portaineree.Endpoint
	name                string
	constraint          string
	newRuleEnabled      bool
	existingRuleEnabled bool
	newRule             *PodSecurityRule
	existingRule        *PodSecurityRule
	constraintFolder    string
}

func (cons *PodSecurityConstraint) Init(userID portainer.UserID, endpoint *portaineree.Endpoint, name string, req *PodSecurityRule, existingRule *PodSecurityRule, constraintFolder string) {
	cons.userID = userID
	cons.endpoint = endpoint
	cons.name = name
	cons.newRule = req
	cons.existingRule = existingRule
	cons.constraintFolder = constraintFolder
}

// deploy the constraint with a maximum 5 times retry, as sometimes need to wait serveral seconds for the template to take effect
func (cons *PodSecurityConstraint) create(kubernetesDeployer portaineree.KubernetesDeployer) error {
	constraint, err := cons.getConstraint()
	if err != nil {
		return err
	}

	cons.constraint = constraint
	//kubctrl apply -f constraint
	log.Info().Str("constraint", cons.constraint).Msg("creating")

	for retry := 0; retry < 5; retry++ {
		_, err = kubernetesDeployer.Deploy(cons.userID, cons.endpoint, []string{cons.constraint}, GateKeeperNameSpace)
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "no matches for kind") {
			log.Info().Str("constraint", cons.constraint).Msg("waiting for template to take effect")

			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	return err
}
func (cons *PodSecurityConstraint) delete(kubernetesDeployer portaineree.KubernetesDeployer) error {
	files, err := cons.getExistingConstraint()
	if err != nil {
		return err
	}

	if _, err := os.Stat(files); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	cons.constraint = files
	//kubctrl delete -f constraint
	_, err = kubernetesDeployer.Remove(cons.userID, cons.endpoint, []string{cons.constraint}, GateKeeperNameSpace)
	if err != nil {
		return err
	}

	os.Remove(cons.constraint)
	return nil
}

// fresh the status of field in k8s
func (cons *PodSecurityConstraint) Fresh(kubernetesDeployer portaineree.KubernetesDeployer) error {
	cons.newRuleEnabled, cons.existingRuleEnabled = cons.getRulesStatus()

	log.Debug().
		Str("constraint", cons.name).
		Bool("request_enabled_status", cons.newRuleEnabled).
		Bool("existing_enabled_status", cons.existingRuleEnabled).
		Msg("updating Pod Security Rule field")

	if cons.newRuleEnabled && cons.existingRuleEnabled {
		//kubctrl delete -f constraint then apply -f new constraint
		err := cons.delete(kubernetesDeployer)
		if err != nil {
			return err
		}

		return cons.create(kubernetesDeployer)
	} else if cons.newRuleEnabled && !cons.existingRuleEnabled {
		return cons.create(kubernetesDeployer)
	} else if !cons.newRuleEnabled && cons.existingRuleEnabled {
		return cons.delete(kubernetesDeployer)
	}

	return nil
}

// generate constraint template yaml file locations according to different Pod Security Rule fields
func (cons *PodSecurityConstraint) getConstraint() (string, error) {
	return createK8SYamlFile(path.Join(cons.constraintFolder, "pod-security-policy", strconv.Itoa(cons.newRule.EndpointID)), cons.name, cons.newRule)
}

// generate constraint yaml files according to different Pod Security Rule fields
func createK8SYamlFile(workDir string, constraint string, rule *PodSecurityRule) (string, error) {
	constraintManifest := PodSecurityConstraintCommon{}
	constraintManifest.APIVersion = "constraints.gatekeeper.sh/v1beta1"
	specKinds := new(struct {
		APIGroups []string `yaml:"apiGroups"`
		Kinds     []string `yaml:"kinds"`
	})
	specKinds.APIGroups = append(specKinds.APIGroups, "")
	specKinds.Kinds = append(specKinds.Kinds, "Pod")
	constraintManifest.Spec.Match.Kinds = append(constraintManifest.Spec.Match.Kinds, *specKinds)

	//Now fill the yaml for each kind of constraint
	constraintManifest.Kind = constraint
	constraintManifest.Metadata.Name = "default_name"
	metadataName, ok := PodSecurityConstraintsMap[constraintManifest.Kind]
	if ok {
		constraintManifest.Metadata.Name = metadataName
	}

	switch constraint {

	case "K8sPSPAppArmor":
		type Parameters struct {
			AllowedProfiles []string `yaml:"allowedProfiles"`
		}

		params := Parameters{}
		params.AllowedProfiles = rule.AppArmour.AppArmourType
		constraintManifest.Spec.Parameters = params
	case "K8sPSPCapabilities":
		type Parameters struct {
			AllowedCapabilities      []string `yaml:"allowedCapabilities"`
			RequiredDropCapabilities []string `yaml:"requiredDropCapabilities"`
		}

		params := Parameters{}
		params.AllowedCapabilities = rule.Capabilities.AllowedCapabilities
		params.RequiredDropCapabilities = rule.Capabilities.RequiredDropCapabilities
		constraintManifest.Spec.Parameters = params
	case "K8sPSPFlexVolumes":
		type AllowedFlexVolumes struct {
			Driver string `yaml:"driver"`
		}

		type Parameters struct {
			AllowedFlexVolumes []AllowedFlexVolumes `yaml:"allowedFlexVolumes"`
		}

		params := Parameters{}
		for _, item := range rule.AllowFlexVolumes.AllowedVolumes {
			vol := AllowedFlexVolumes{}
			vol.Driver = item
			params.AllowedFlexVolumes = append(params.AllowedFlexVolumes, vol)
		}

		constraintManifest.Spec.Parameters = params
	case "K8sPSPForbiddenSysctls":
		type Parameters struct {
			ForbiddenSysctls []string `yaml:"forbiddenSysctls"`
		}

		params := Parameters{}
		params.ForbiddenSysctls = rule.ForbiddenSysctlsList.RequiredDropCapabilities
		constraintManifest.Spec.Parameters = params
	case "K8sPSPHostFilesystem":
		type AllowedHostPaths struct {
			ReadOnly   bool   `yaml:"readOnly"`
			PathPrefix string `yaml:"pathPrefix"`
		}

		type Parameters struct {
			AllowedHostPaths []AllowedHostPaths `yaml:"allowedHostPaths"`
		}

		params := Parameters{}
		for _, path := range rule.HostFilesystem.AllowedPaths {
			allowedPath := AllowedHostPaths{}
			allowedPath.PathPrefix = path.PathPrefix
			allowedPath.ReadOnly = path.Readonly
			params.AllowedHostPaths = append(params.AllowedHostPaths, allowedPath)
		}

		constraintManifest.Spec.Parameters = params

	case "K8sPSPHostNetworkingPorts":
		type Parameters struct {
			Min         int  `yaml:"min"`
			Max         int  `yaml:"max"`
			HostNetwork bool `yaml:"hostNetwork"`
		}

		params := Parameters{}
		params.HostNetwork = rule.HostNetworkingPorts.HostNetwork
		params.Max = rule.HostNetworkingPorts.Max
		params.Min = rule.HostNetworkingPorts.Min
		constraintManifest.Spec.Parameters = params

	case "K8sPSPPrivilegedContainer":
		constraintManifest.Spec.Match.ExcludedNamespaces = append(constraintManifest.Spec.Match.ExcludedNamespaces, "kube-system")

	case "K8sPSPProcMount":
		constraintManifest.Metadata.Name = "psp-proc-mount"

		type Parameters struct {
			ProcMount string `yaml:"procMount"`
		}

		params := Parameters{}
		params.ProcMount = rule.AllowProcMount.ProcMountType
		constraintManifest.Spec.Parameters = params
	case "K8sPSPSeccomp":
		constraintManifest.Metadata.Name = "psp-seccomp"

		type Parameters struct {
			AllowedProfiles []string `yaml:"allowedProfiles"`
		}

		params := Parameters{}
		params.AllowedProfiles = rule.SecComp.SecCompType
		constraintManifest.Spec.Parameters = params
	case "K8sPSPSELinuxV2":
		type AllowedSELinuxOptions struct {
			Level string `yaml:"level"`
			Role  string `yaml:"role"`
			Type  string `yaml:"type"`
			User  string `yaml:"user"`
		}

		type Parameters struct {
			AllowedSELinuxOptions []AllowedSELinuxOptions `yaml:"allowedSELinuxOptions"`
		}

		params := Parameters{}
		for _, item := range rule.Selinux.AllowedCapabilities {
			option := AllowedSELinuxOptions{}
			option.Level = item.Level
			option.Role = item.Role
			option.Type = item.Type
			option.User = item.User
			params.AllowedSELinuxOptions = append(params.AllowedSELinuxOptions, option)
		}

		constraintManifest.Spec.Parameters = params
	case "K8sPSPAllowedUsers":
		type Ranges struct {
			Min int `yaml:"min"`
			Max int `yaml:"max"`
		}

		// RunAsUser
		type RunAsSomebody struct {
			Rule   string   `yaml:"rule"`
			Ranges []Ranges `yaml:"ranges"`
		}

		// Parameters
		type Parameters struct {
			RunAsUser          RunAsSomebody `yaml:"runAsUser"`
			RunAsGroup         RunAsSomebody `yaml:"runAsGroup"`
			SupplementalGroups RunAsSomebody `yaml:"supplementalGroups"`
			FsGroup            RunAsSomebody `yaml:"fsGroup"`
		}
		params := Parameters{}

		//handle runAsUser
		params.RunAsUser.Rule = string(rule.Users.RunAsUser.Type)
		if rule.Users.RunAsUser.Type == RunAsUserStrategyMustRunAs {
			//handle range list
			for _, item := range rule.Users.RunAsUser.Idrange {
				rg := Ranges{}
				rg.Max = item.Max
				rg.Min = item.Min
				params.RunAsUser.Ranges = append(params.RunAsUser.Ranges, rg)
			}

		}

		//handle RunAsGroup
		params.RunAsGroup.Rule = string(rule.Users.RunAsGroup.Type)
		if rule.Users.RunAsGroup.Type == RunAsGroupStrategyMustRunAs || rule.Users.RunAsGroup.Type == RunAsGroupStrategyMayRunAs {
			//handle range list
			for _, item := range rule.Users.RunAsGroup.Idrange {
				rg := Ranges{}
				rg.Max = item.Max
				rg.Min = item.Min
				params.RunAsGroup.Ranges = append(params.RunAsGroup.Ranges, rg)
			}
		}

		//handle SupplementalGroups
		params.SupplementalGroups.Rule = string(rule.Users.SupplementalGroups.Type)
		if rule.Users.SupplementalGroups.Type == SupplementalGroupsStrategyMustRunAs || rule.Users.SupplementalGroups.Type == SupplementalGroupsStrategyMayRunAs {
			//handle range list
			for _, item := range rule.Users.SupplementalGroups.Idrange {
				rg := Ranges{}
				rg.Max = item.Max
				rg.Min = item.Min
				params.SupplementalGroups.Ranges = append(params.SupplementalGroups.Ranges, rg)
			}
		}

		params.FsGroup.Rule = string(rule.Users.FsGroups.Type)
		if rule.Users.FsGroups.Type == FSGroupStrategyMustRunAs || rule.Users.FsGroups.Type == FSGroupStrategyMayRunAs {
			for _, item := range rule.Users.FsGroups.Idrange {
				rg := Ranges{}
				rg.Max = item.Max
				rg.Min = item.Min
				params.FsGroup.Ranges = append(params.FsGroup.Ranges, rg)
			}
		}

		constraintManifest.Spec.Parameters = params

	case "K8sPSPVolumeTypes":
		type Parameters struct {
			Volumes []string `yaml:"volumes"`
		}

		params := Parameters{}
		for _, vols := range rule.VolumeTypes.AllowedTypes {
			if string(vols) == "*" {
				params.Volumes = []string{"*"}
				break
			}

			params.Volumes = append(params.Volumes, string(vols))
		}

		constraintManifest.Spec.Parameters = params

	default:
		//K8sPSPAllowPrivilegeEscalationContainer K8sPSPHostNamespace K8sPSPReadOnlyRootFilesystem have no extra parameters
	}

	yamlData, err := yaml.Marshal(&constraintManifest)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error while Marshaling. %v", err))
	}

	manifestFilePath := filesystem.JoinPaths(workDir, constraint+".yaml")

	err = filesystem.WriteToFile(manifestFilePath, yamlData)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create constraint yaml file: %v", manifestFilePath)
	}

	return manifestFilePath, nil
}

func (cons *PodSecurityConstraint) getExistingConstraint() (string, error) {
	return path.Join(cons.constraintFolder, "pod-security-policy", strconv.Itoa(cons.newRule.EndpointID), cons.name+".yaml"), nil
}

// check if the field needs to be created/updated/deleted by comparing the request value and database value
func (cons *PodSecurityConstraint) getRulesStatus() (bool, bool) {
	switch cons.name {
	case "K8sPSPAllowPrivilegeEscalationContainer":
		return cons.newRule.AllowPrivilegeEscalation.Enabled, cons.existingRule.AllowPrivilegeEscalation.Enabled
	case "K8sPSPAppArmor":
		return cons.newRule.AppArmour.Enabled, cons.existingRule.AppArmour.Enabled
	case "K8sPSPCapabilities":
		return cons.newRule.Capabilities.Enabled, cons.existingRule.Capabilities.Enabled
	case "K8sPSPFlexVolumes":
		return cons.newRule.AllowFlexVolumes.Enabled, cons.existingRule.AllowFlexVolumes.Enabled
	case "K8sPSPForbiddenSysctls":
		return cons.newRule.ForbiddenSysctlsList.Enabled, cons.existingRule.ForbiddenSysctlsList.Enabled
	case "K8sPSPHostFilesystem":
		return cons.newRule.HostFilesystem.Enabled, cons.existingRule.HostFilesystem.Enabled
	case "K8sPSPHostNamespace":
		return cons.newRule.HostNamespaces.Enabled, cons.existingRule.HostNamespaces.Enabled
	case "K8sPSPHostNetworkingPorts":
		return cons.newRule.HostNetworkingPorts.Enabled, cons.existingRule.HostNetworkingPorts.Enabled
	case "K8sPSPPrivilegedContainer":
		return cons.newRule.PrivilegedContainers.Enabled, cons.existingRule.PrivilegedContainers.Enabled
	case "K8sPSPProcMount":
		return cons.newRule.AllowProcMount.Enabled, cons.existingRule.AllowProcMount.Enabled
	case "K8sPSPReadOnlyRootFilesystem":
		return cons.newRule.ReadOnlyRootFileSystem.Enabled, cons.existingRule.ReadOnlyRootFileSystem.Enabled
	case "K8sPSPSeccomp":
		return cons.newRule.SecComp.Enabled, cons.existingRule.SecComp.Enabled
	case "K8sPSPSELinuxV2":
		return cons.newRule.Selinux.Enabled, cons.existingRule.Selinux.Enabled
	case "K8sPSPAllowedUsers":
		return cons.newRule.Users.Enabled, cons.existingRule.Users.Enabled
	case "K8sPSPVolumeTypes":
		return cons.newRule.VolumeTypes.Enabled, cons.existingRule.VolumeTypes.Enabled
	default:
		return false, false
	}
}
