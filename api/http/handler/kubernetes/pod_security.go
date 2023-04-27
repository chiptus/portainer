package kubernetes

import (
	"encoding/json"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/kubernetes/podsecurity"

	"github.com/rs/zerolog/log"
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
// @param environmentId path int true "Environment identifier"
// @success 200 {object} podsecurity.PodSecurityRule "Success"
// @failure 400 "Invalid request"
// @router /kubernetes/{environmentId}/opa [get]
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
// @param environmentId path int true "Environment(Endpoint) identifier"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 404 "Pod Security Rule not found"
// @failure 500 "Server error"
// @router /kubernetes/{environmentId}/opa [put]
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

	gateKeeper := podsecurity.NewGateKeeper(
		handler.KubernetesDeployer,
		handler.baseFileDir,
	)

	if !requestRule.Enabled {
		err = gateKeeper.Remove(tokenData.ID, endpoint)
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
		err = gateKeeper.Deploy(tokenData.ID, endpoint)
		if err != nil {
			return httperror.InternalServerError("failed to deploy kubernetes gatekeeper", err)
		}

		err := checkGatekeeperStatus(handler, endpoint, r)
		if err != nil {
			return httperror.InternalServerError("unable to get gatekeeper status", err)
		}

		//2. deploy gatekeeper excluded namespaces
		err = gateKeeper.DeployExcludedNamespaces(tokenData.ID, endpoint)
		if err != nil {
			return httperror.InternalServerError("failed to deploy excluded namespaces for gatekeeper", err)
		}

		//3.deploy gatekeeper constrainttemplate
		err = gateKeeper.DeployPodSecurityConstraints(tokenData.ID, endpoint)
		if err != nil {
			return httperror.InternalServerError("failed to deploy pod security constraints for gatekeeper", err)
		}

		err = handler.DataStore.PodSecurity().Create(existedRule)
		if err != nil {
			return httperror.InternalServerError("Unable to add the pod security rule to the database", err)
		}
	}

	//4.deploy gatekeeper constraint yaml files
	for name := range podsecurity.PodSecurityConstraintsMap {
		rulename := name
		constraint := podsecurity.PodSecurityConstraint{}
		constraint.Init(tokenData.ID, endpoint, rulename, requestRule, existedRule, handler.fileService.GetDatastorePath())

		if err := constraint.Fresh(handler.KubernetesDeployer); err != nil {
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
var checkGatekeeperStatus = func(handler *Handler, endpoint *portaineree.Endpoint, r *http.Request) error {
	cli, err := handler.KubernetesClientFactory.CreateClient(endpoint)
	if err != nil {
		return err
	}

	return podsecurity.WaitForOpaReady(r.Context(), cli)
}
