package endpoints

import (
	"net/http"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
)

// @id endpointRegistriesList
// @summary List Registries on environment
// @description List all registries based on the current user authorizations in current environment.
// @description **Access policy**: authenticated
// @tags endpoints
// @param namespace query string false "required if kubernetes environment, will show registries by namespace"
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {array} portaineree.Registry "Success"
// @failure 500 "Server error"
// @router /endpoints/{id}/registries [get]
func (handler *Handler) endpointRegistriesList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve info from request context", Err: err}
	}

	user, err := handler.dataStore.User().User(securityContext.UserID)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve user from the database", Err: err}
	}

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid environment identifier route variable", Err: err}
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	}

	isAdminOrEndpointAdmin, err := security.IsAdminOrEndpointAdmin(r, handler.dataStore, endpoint.ID)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check user role", Err: err}
	}

	registries, err := handler.dataStore.Registry().Registries()
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve registries from the database", Err: err}
	}

	if endpointutils.IsKubernetesEndpoint(endpoint) {
		namespace, _ := request.RetrieveQueryParameter(r, "namespace", true)

		if namespace == "" && !isAdminOrEndpointAdmin {
			return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Missing namespace query parameter", Err: errors.New("missing namespace query parameter")}
		}

		if namespace != "" {
			authorized, err := handler.isNamespaceAuthorized(endpoint, namespace, user.ID, securityContext.UserMemberships, isAdminOrEndpointAdmin)
			if err != nil {
				return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to check for namespace authorization", Err: err}
			}

			if !authorized {
				return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "User is not authorized to use namespace", Err: errors.New("user is not authorized to use namespace")}
			}

			registries = filterRegistriesByNamespace(registries, endpoint.ID, namespace)
		}

	} else if !isAdminOrEndpointAdmin {
		registries = security.FilterRegistries(registries, user, securityContext.UserMemberships, endpoint.ID)
	}

	for idx := range registries {
		hideRegistryFields(&registries[idx], !isAdminOrEndpointAdmin)
	}

	return response.JSON(w, registries)
}

func (handler *Handler) isNamespaceAuthorized(endpoint *portaineree.Endpoint, namespace string, userId portaineree.UserID, memberships []portaineree.TeamMembership, isAdminOrEndpointAdmin bool) (bool, error) {
	if isAdminOrEndpointAdmin || namespace == "" {
		return true, nil
	}

	if !endpoint.Kubernetes.Configuration.RestrictDefaultNamespace && namespace == "default" {
		return true, nil
	}

	kcl, err := handler.K8sClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return false, errors.Wrap(err, "unable to retrieve kubernetes client")
	}

	accessPolicies, err := kcl.GetNamespaceAccessPolicies()
	if err != nil {
		return false, errors.Wrap(err, "unable to retrieve environment's namespaces policies")
	}

	namespacePolicy, ok := accessPolicies[namespace]
	if !ok {
		return false, nil
	}

	return security.AuthorizedAccess(userId, memberships, namespacePolicy.UserAccessPolicies, namespacePolicy.TeamAccessPolicies), nil
}

func filterRegistriesByNamespace(registries []portaineree.Registry, endpointId portaineree.EndpointID, namespace string) []portaineree.Registry {
	if namespace == "" {
		return registries
	}

	filteredRegistries := []portaineree.Registry{}

	for _, registry := range registries {
		for _, authorizedNamespace := range registry.RegistryAccesses[endpointId].Namespaces {
			if authorizedNamespace == namespace {
				filteredRegistries = append(filteredRegistries, registry)
			}
		}
	}

	return filteredRegistries
}

func hideRegistryFields(registry *portaineree.Registry, hideAccesses bool) {
	registry.Password = ""
	registry.ManagementConfiguration = nil
	if hideAccesses {
		registry.RegistryAccesses = nil
	}
}
