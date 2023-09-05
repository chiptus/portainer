package endpoints

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer/pkg/featureflags"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
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
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	registryID, err := request.RetrieveNumericRouteVariableValue(r, "registryId")
	if err != nil {
		return httperror.BadRequest("Invalid registry identifier route variable", err)
	}

	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		err = handler.updateRegistryAccess(handler.DataStore, r, portaineree.EndpointID(endpointID), portaineree.RegistryID(registryID))
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			return handler.updateRegistryAccess(tx, r, portaineree.EndpointID(endpointID), portaineree.RegistryID(registryID))
		})
	}

	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.Empty(w)
}

func (handler *Handler) updateRegistryAccess(tx dataservices.DataStoreTx, r *http.Request, endpointID portaineree.EndpointID, registryID portaineree.RegistryID) error {
	endpoint, err := tx.Endpoint().Endpoint(endpointID)
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	isAdminOrEndpointAdmin, err := security.IsAdminOrEndpointAdmin(r, tx, endpointID)
	if err != nil {
		return httperror.InternalServerError("Unable to check user role", err)
	}
	if !isAdminOrEndpointAdmin {
		return httperror.Forbidden("User is not authorized", err)
	}

	registry, err := tx.Registry().Read(registryID)
	if tx.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	var payload registryAccessPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
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
			return httperror.InternalServerError("Unable to update kube access policies", err)
		}

		registryAccess.Namespaces = payload.Namespaces
	} else {
		registryAccess.UserAccessPolicies = payload.UserAccessPolicies
		registryAccess.TeamAccessPolicies = payload.TeamAccessPolicies
	}

	registry.RegistryAccesses[portaineree.EndpointID(endpointID)] = registryAccess

	return tx.Registry().Update(registry.ID, registry)
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
