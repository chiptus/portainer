package kubernetes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/kubernetes/podsecurity"
	"github.com/portainer/portainer/api/filesystem"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// @id getKubernetesPodSecurityRule
// @summary Get Pod Security Rule within k8s cluster, if not found, the frontend will create a default
// @description Get Pod Security Rule within k8s cluster
// @description **Access policy**: authenticated
// @tags kubernetes
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {object} podsecurity.PodSecurityRule "Success"
// @failure 400 "Invalid request"
// @router /kubernetes/{id}/opa [get]
func (handler *Handler) getK8sPodSecurityRule(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}
	securityRule, err := handler.DataStore.PodSecurity().PodSecurityByEndpointID(int(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		securityRule := &podsecurity.PodSecurityRule{}

		securityRule.Users = podsecurity.PodSecurityUsers{}
		securityRule.Users.RunAsUser = podsecurity.PodSecurityRunAsUser{
			Type:    podsecurity.RunAsUserStrategyMustRunAs,
			Idrange: make([]podsecurity.PodSecurityIdrange, 0),
		}
		securityRule.Users.RunAsGroup = podsecurity.PodSecurityRunAsGroup{
			Type:    podsecurity.RunAsGroupStrategyMustRunAs,
			Idrange: make([]podsecurity.PodSecurityIdrange, 0),
		}
		securityRule.Users.SupplementalGroups = podsecurity.PodSecuritySupplementalGroups{
			Type:    podsecurity.SupplementalGroupsStrategyMustRunAs,
			Idrange: make([]podsecurity.PodSecurityIdrange, 0),
		}
		securityRule.Users.FsGroups = podsecurity.PodSecurityFsGroups{
			Type:    podsecurity.FSGroupStrategyMustRunAs,
			Idrange: make([]podsecurity.PodSecurityIdrange, 0),
		}
		securityRule.SecComp = podsecurity.PodSecuritySecComp{
			Enabled:     false,
			SecCompType: make([]string, 0),
		}
		securityRule.VolumeTypes = podsecurity.PodSecurityVolumeTypes{
			Enabled:      false,
			AllowedTypes: make([]podsecurity.FSType, 0),
		}
		securityRule.HostFilesystem = podsecurity.PodSecurityHostFilesystem{
			Enabled:      false,
			AllowedPaths: make([]podsecurity.PodSecurityAllowedPaths, 0),
		}
		securityRule.AllowFlexVolumes = podsecurity.PodSecurityAllowFlexVolumes{
			Enabled:        false,
			AllowedVolumes: make([]string, 0),
		}
		securityRule.Capabilities = podsecurity.PodSecurityCapabilities{
			Enabled:                  false,
			AllowedCapabilities:      make([]string, 0),
			RequiredDropCapabilities: make([]string, 0),
		}
		securityRule.Selinux = podsecurity.PodSecuritySelinux{
			Enabled:             false,
			AllowedCapabilities: make([]podsecurity.PodSecurityAllowedCapabilities, 0),
		}
		securityRule.AllowProcMount = podsecurity.PodSecurityAllowProcMount{
			Enabled:       false,
			ProcMountType: "Default",
		}
		securityRule.AppArmour = podsecurity.PodSecurityAppArmour{
			Enabled:       false,
			AppArmourType: make([]string, 0),
		}
		securityRule.ForbiddenSysctlsList = podsecurity.PodSecurityForbiddenSysctlsList{
			Enabled:                  false,
			RequiredDropCapabilities: make([]string, 0),
		}

		return response.JSON(w, securityRule)
	} else if err != nil {
		return httperror.InternalServerError("Unable to retrieve pod security rule from the database", err)
	}
	return response.JSON(w, securityRule)
}

// @id updateK8sPodSecurityRule
// @summary Update Pod Security Rule within k8s cluster
// @description Update Pod Security Rule within k8s cluster
// @description **Access policy**: authenticated
// @tags kubernetes
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 404 "Pod Security Rule not found"
// @failure 500 "Server error"
// @router /kubernetes/{endpointId}/opa [put]
func (handler *Handler) updateK8sPodSecurityRule(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	handler.opaOperationMutex.Lock()
	defer handler.opaOperationMutex.Unlock()

	requestRule := &podsecurity.PodSecurityRule{}
	err := json.NewDecoder(r.Body).Decode(requestRule)
	if err != nil {
		return httperror.BadRequest("cannot parse request body", err)
	}
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}
	requestRule.EndpointID = endpointID

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}
	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve user details from authentication token", err)
	}

	gatekeeperManifest := path.Join(handler.baseFileDir, "pod-security-policy", podsecurity.GateKeeperFile)
	if !requestRule.Enabled {
		_, err = handler.KubernetesDeployer.Remove(tokenData.ID, endpoint, []string{gatekeeperManifest}, podsecurity.GateKeeperNameSpace)
		if err != nil {
			return httperror.InternalServerError("failed to remove kubernetes gatekeeper", err)
		}

		err = handler.DataStore.PodSecurity().DeletePodSecurityRule(endpointID)
		if err != nil {
			log.Error().Err(err).Int("endpoint_id", endpointID).Msg("failed to delete pod security rule")
		}

		return response.JSON(w, requestRule)
	}

	isNewPodSecurity := false
	existedRule, err := handler.DataStore.PodSecurity().PodSecurityByEndpointID(endpointID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		isNewPodSecurity = true
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a pod security rule with the specified identifier inside the database", err)
	}
	if isNewPodSecurity {
		existedRule = &podsecurity.PodSecurityRule{}
		existedRule.EndpointID = endpointID

		//1.deploy gatekeeper
		_, err = handler.KubernetesDeployer.Deploy(tokenData.ID, endpoint, []string{gatekeeperManifest}, podsecurity.GateKeeperNameSpace)
		if err != nil {
			log.Error().Msg("failed to deploy kubernetes gatekeeper, remove installed files")
			handler.KubernetesDeployer.Remove(tokenData.ID, endpoint, []string{gatekeeperManifest}, podsecurity.GateKeeperNameSpace)
			return httperror.InternalServerError("failed to deploy kubernetes gatekeeper", err)
		}

		err := checkGetekeeperStatus(handler, endpoint, r)
		if err != nil {
			return httperror.InternalServerError("unable to get gatekeeper status", err)
		}

		//2. deploy gatekeeper excluded namespaces
		gatekeeperExcludedNamespacesManifest := path.Join(handler.baseFileDir, "pod-security-policy", podsecurity.GateKeeperExcludedNamespacesFile)
		_, err = handler.KubernetesDeployer.Deploy(tokenData.ID, endpoint, []string{gatekeeperExcludedNamespacesManifest}, podsecurity.GateKeeperNameSpace)
		if err != nil {
			log.Error().Msg("failed to apply kubernetes gatekeeper namespace exclusions")
			handler.KubernetesDeployer.Remove(tokenData.ID, endpoint, []string{gatekeeperManifest}, podsecurity.GateKeeperNameSpace)
			return httperror.InternalServerError("failed to deploy kubernetes gatekeeper", err)
		}

		//2.deploy gatekeeper constrainttemplate
		for _, v := range podsecurity.PodSecurityConstraintsMap {
			log.Info().Str("template", v).Msg("deploying constraint template")

			_, err := handler.KubernetesDeployer.Deploy(tokenData.ID, endpoint, []string{path.Join(handler.baseFileDir, "pod-security-policy", v, "template.yaml")}, podsecurity.GateKeeperNameSpace)
			if err != nil {
				log.Error().Str("template", v).Err(err).Msg("unable to deploy")

				return httperror.InternalServerError("Unable to apply the pod security rule templates to the system", err)
			}

			log.Info().Str("template", v).Msg("successfully deployed")
		}

		err = handler.DataStore.PodSecurity().Create(existedRule)
		if err != nil {
			return httperror.InternalServerError("Unable to add the pod security rule to the database", err)
		}
	}

	//3.deploy gatekeeper constraint yaml files
	for name := range podsecurity.PodSecurityConstraintsMap {
		rulename := name
		constraint := PodSecurityConstraint{}
		constraint.init(tokenData.ID, endpoint, rulename, requestRule, existedRule, handler.fileService.GetDatastorePath(), handler)

		if err := constraint.fresh(handler); err != nil {
			log.Error().Str("rule", rulename).Err(err).Msg("unable to deploy constraint")

			return httperror.InternalServerError("Unable to apply the pod security rule constraints to the system", err)
		}

		log.Info().Str("rule", rulename).Msg("successfully deployed constraint")
	}

	err = handler.DataStore.PodSecurity().UpdatePodSecurityRule(existedRule.EndpointID, requestRule)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the pod security rule changes inside the database", err)
	}

	return response.JSON(w, requestRule)
}

// check the gatekeeper pod status
var checkGetekeeperStatus = func(handler *Handler, endpoint *portaineree.Endpoint, r *http.Request) error {
	cli, err := handler.KubernetesClientFactory.CreateClient(endpoint)
	if err != nil {
		return err
	}

	return podsecurity.WaitForOpaReady(r.Context(), cli)
}

type PodSecurityConstraint struct {
	userID              portaineree.UserID
	endpoint            *portaineree.Endpoint
	name                string
	constraint          string
	newRuleEnabled      bool
	existingRuleEnabled bool
	newRule             *podsecurity.PodSecurityRule
	existingRule        *podsecurity.PodSecurityRule
	templateFolder      string
	constraintFolder    string
}

func (cons *PodSecurityConstraint) init(userID portaineree.UserID, endpoint *portaineree.Endpoint, name string, req *podsecurity.PodSecurityRule, existingRule *podsecurity.PodSecurityRule, constraintFolder string, handler *Handler) {
	cons.userID = userID
	cons.endpoint = endpoint
	cons.name = name
	cons.newRule = req
	cons.existingRule = existingRule
	cons.constraintFolder = constraintFolder
	cons.templateFolder = handler.baseFileDir
}

// deploy the constraint with a maximum 5 times retry, as sometimes need to wait serveral seconds for the template to take effect
func (cons *PodSecurityConstraint) create(handler *Handler) error {
	constraint, err := cons.getConstraint()
	if err != nil {
		return err
	}

	cons.constraint = constraint
	//kubctrl apply -f constraint
	log.Info().Str("constraint", cons.constraint).Msg("creating")

	for retry := 0; retry < 5; retry++ {
		_, err = handler.KubernetesDeployer.Deploy(cons.userID, cons.endpoint, []string{cons.constraint}, podsecurity.GateKeeperNameSpace)
		if err != nil && strings.Contains(strings.ToLower(err.Error()), "no matches for kind") {
			log.Info().Str("constraint", cons.constraint).Msg("waiting for template to take effect")

			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	return err
}
func (cons *PodSecurityConstraint) delete(handler *Handler) error {
	files, err := cons.getExistingConstraint()
	if err != nil {
		return err
	}

	if _, err := os.Stat(files); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	cons.constraint = files
	//kubctrl delete -f constraint
	_, err = handler.KubernetesDeployer.Remove(cons.userID, cons.endpoint, []string{cons.constraint}, podsecurity.GateKeeperNameSpace)
	if err != nil {
		return err
	}

	os.Remove(cons.constraint)
	return nil
}

// fresh the status of field in k8s
func (cons *PodSecurityConstraint) fresh(handler *Handler) error {
	log.Debug().
		Str("constraint", cons.name).
		Bool("request_enabled_status", cons.newRuleEnabled).
		Bool("existing_enabled_status", cons.existingRuleEnabled).
		Msg("updating Pod Security Rule field")

	if cons.newRuleEnabled && cons.existingRuleEnabled {
		//kubctrl delete -f constraint then apply -f new constraint
		err := cons.delete(handler)
		if err != nil {
			return err
		}

		return cons.create(handler)
	} else if cons.newRuleEnabled && !cons.existingRuleEnabled {
		return cons.create(handler)
	} else if !cons.newRuleEnabled && cons.existingRuleEnabled {
		return cons.delete(handler)
	}

	return nil
}

// generate constraint template yaml file locations according to different Pod Security Rule fields
func (cons *PodSecurityConstraint) getConstraint() (string, error) {
	return createK8SYamlFile(path.Join(cons.constraintFolder, "pod-security-policy", strconv.Itoa(cons.newRule.EndpointID)), cons.name, cons.newRule)
}

// generate constraint yaml files according to different Pod Security Rule fields
func createK8SYamlFile(workDir string, constraint string, rule *podsecurity.PodSecurityRule) (string, error) {
	constraintManifest := podsecurity.PodSecurityConstraintCommon{}
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
	metadataName, ok := podsecurity.PodSecurityConstraintsMap[constraintManifest.Kind]
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
		if rule.Users.RunAsUser.Type == podsecurity.RunAsUserStrategyMustRunAs {
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
		if rule.Users.RunAsGroup.Type == podsecurity.RunAsGroupStrategyMustRunAs || rule.Users.RunAsGroup.Type == podsecurity.RunAsGroupStrategyMayRunAs {
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
		if rule.Users.SupplementalGroups.Type == podsecurity.SupplementalGroupsStrategyMustRunAs || rule.Users.SupplementalGroups.Type == podsecurity.SupplementalGroupsStrategyMayRunAs {
			//handle range list
			for _, item := range rule.Users.SupplementalGroups.Idrange {
				rg := Ranges{}
				rg.Max = item.Max
				rg.Min = item.Min
				params.SupplementalGroups.Ranges = append(params.SupplementalGroups.Ranges, rg)
			}
		}

		params.FsGroup.Rule = string(rule.Users.FsGroups.Type)
		if rule.Users.FsGroups.Type == podsecurity.FSGroupStrategyMustRunAs || rule.Users.FsGroups.Type == podsecurity.FSGroupStrategyMayRunAs {
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
