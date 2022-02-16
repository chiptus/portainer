package endpoints

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

type registryAccessPayload struct {
	UserAccessPolicies portaineree.UserAccessPolicies
	TeamAccessPolicies portaineree.TeamAccessPolicies
	Namespaces         []string
}

func (payload *registryAccessPayload) Validate(r *http.Request) error {
	return nil
}

// @id endpointRegistryAccess
// @summary update registry access for environment
// @description **Access policy**: authenticated
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @param registryId path int true "Registry identifier"
// @param body body registryAccessPayload true "details"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Endpoint not found"
// @failure 500 "Server error"
// @router /endpoints/{id}/registries/{registryId} [put]
func (handler *Handler) endpointRegistryAccess(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid environment identifier route variable", Err: err}
	}

	registryID, err := request.RetrieveNumericRouteVariableValue(r, "registryId")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid registry identifier route variable", Err: err}
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access environment", err}
	}

	isAdminOrEndpointAdmin, err := security.IsAdminOrEndpointAdmin(r, handler.dataStore, portaineree.EndpointID(endpointID))
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check user role", Err: err}
	}
	if !isAdminOrEndpointAdmin {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "User is not authorized", Err: err}
	}

	registry, err := handler.dataStore.Registry().Registry(portaineree.RegistryID(registryID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	}

	var payload registryAccessPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	if registry.RegistryAccesses == nil {
		registry.RegistryAccesses = portaineree.RegistryAccesses{}
	}

	if _, ok := registry.RegistryAccesses[endpoint.ID]; !ok {
		registry.RegistryAccesses[endpoint.ID] = portaineree.RegistryAccessPolicies{}
	}

	registryAccess := registry.RegistryAccesses[endpoint.ID]

	if endpoint.Type == portaineree.KubernetesLocalEnvironment || endpoint.Type == portaineree.AgentOnKubernetesEnvironment || endpoint.Type == portaineree.EdgeAgentOnKubernetesEnvironment {
		err := handler.updateKubeAccess(endpoint, registry, registryAccess.Namespaces, payload.Namespaces)
		if err != nil {
			return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to update kube access policies", Err: err}
		}

		registryAccess.Namespaces = payload.Namespaces
	} else {
		registryAccess.UserAccessPolicies = payload.UserAccessPolicies
		registryAccess.TeamAccessPolicies = payload.TeamAccessPolicies
	}

	registry.RegistryAccesses[portaineree.EndpointID(endpointID)] = registryAccess

	handler.dataStore.Registry().UpdateRegistry(registry.ID, registry)

	return response.Empty(w)
}

func (handler *Handler) updateKubeAccess(endpoint *portaineree.Endpoint, registry *portaineree.Registry, oldNamespaces, newNamespaces []string) error {
	oldNamespacesSet := toSet(oldNamespaces)
	newNamespacesSet := toSet(newNamespaces)

	namespacesToRemove := setDifference(oldNamespacesSet, newNamespacesSet)
	namespacesToAdd := setDifference(newNamespacesSet, oldNamespacesSet)

	cli, err := handler.K8sClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return err
	}

	for namespace := range namespacesToRemove {
		err := cli.DeleteRegistrySecret(registry, namespace)
		if err != nil {
			return err
		}
	}

	for namespace := range namespacesToAdd {
		err := cli.CreateRegistrySecret(registry, namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

type stringSet map[string]bool

func toSet(list []string) stringSet {
	set := stringSet{}
	for _, el := range list {
		set[el] = true
	}
	return set
}

// setDifference returns the set difference tagsA - tagsB
func setDifference(setA stringSet, setB stringSet) stringSet {
	set := stringSet{}

	for el := range setA {
		if !setB[el] {
			set[el] = true
		}
	}

	return set
}
