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

	registries, handleError := handler.filterRegistriesByAccess(r, registries, endpoint, user, securityContext.UserMemberships)
	if handleError != nil {
		return handleError
	}

	for idx := range registries {
		hideRegistryFields(&registries[idx], !isAdminOrEndpointAdmin)
	}

	return response.JSON(w, registries)
}

func (handler *Handler) filterRegistriesByAccess(r *http.Request, registries []portaineree.Registry, endpoint *portaineree.Endpoint, user *portaineree.User, memberships []portaineree.TeamMembership) ([]portaineree.Registry, *httperror.HandlerError) {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return security.FilterRegistries(registries, user, memberships, endpoint.ID), nil
	}

	return handler.filterKubernetesEndpointRegistries(r, registries, endpoint, user, memberships)
}

func (handler *Handler) filterKubernetesEndpointRegistries(r *http.Request, registries []portaineree.Registry, endpoint *portaineree.Endpoint, user *portaineree.User, memberships []portaineree.TeamMembership) ([]portaineree.Registry, *httperror.HandlerError) {
	namespaceParam, _ := request.RetrieveQueryParameter(r, "namespace", true)
	isAdminOrEndpointAdmin, err := security.IsAdminOrEndpointAdmin(r, handler.dataStore, endpoint.ID)
	if err != nil {
		return nil, &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check user role", Err: err}
	}

	if namespaceParam != "" {
		authorized, err := handler.isNamespaceAuthorized(endpoint, namespaceParam, user.ID, memberships, isAdminOrEndpointAdmin)
		if err != nil {
			return nil, &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to check for namespace authorization", Err: err}
		}
		if !authorized {
			return nil, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "User is not authorized to use namespace", Err: errors.New("user is not authorized to use namespace")}
		}

		return filterRegistriesByNamespaces(registries, endpoint.ID, []string{namespaceParam}), nil
	}

	if isAdminOrEndpointAdmin {
		return registries, nil
	}

	return handler.filterKubernetesRegistriesByUserRole(r, registries, endpoint, user)
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

func filterRegistriesByNamespaces(registries []portaineree.Registry, endpointId portaineree.EndpointID, namespaces []string) []portaineree.Registry {
	filteredRegistries := []portaineree.Registry{}

	for _, registry := range registries {
		if registryAccessPoliciesContainsNamespace(registry.RegistryAccesses[endpointId], namespaces) {
			filteredRegistries = append(filteredRegistries, registry)
		}
	}

	return filteredRegistries
}

func registryAccessPoliciesContainsNamespace(registryAccess portaineree.RegistryAccessPolicies, namespaces []string) bool {
	for _, authorizedNamespace := range registryAccess.Namespaces {
		for _, namespace := range namespaces {
			if namespace == authorizedNamespace {
				return true
			}
		}
	}
	return false
}

func (handler *Handler) filterKubernetesRegistriesByUserRole(r *http.Request, registries []portaineree.Registry, endpoint *portaineree.Endpoint, user *portaineree.User) ([]portaineree.Registry, *httperror.HandlerError) {
	err := handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err == security.ErrAuthorizationRequired {
		return nil, &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "User is not authorized", Err: errors.New("missing namespace query parameter")}
	}
	if err != nil {
		return nil, &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve info from request context", Err: err}
	}

	userNamespaces, err := handler.userNamespaces(endpoint, user)
	if err != nil {
		return nil, &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "unable to retrieve user namespaces", Err: err}
	}

	return filterRegistriesByNamespaces(registries, endpoint.ID, userNamespaces), nil
}

func (handler *Handler) userNamespaces(endpoint *portaineree.Endpoint, user *portaineree.User) ([]string, error) {
	kcl, err := handler.K8sClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return nil, err
	}
	namespaceAuthorizations, err := handler.AuthorizationService.GetNamespaceAuthorizations(int(user.ID), *endpoint, kcl)
	if err != nil {
		return nil, err
	}

	var userNamespaces []string
	for userNamespace := range namespaceAuthorizations {
		userNamespaces = append(userNamespaces, userNamespace)
	}
	return userNamespaces, nil
}

func hideRegistryFields(registry *portaineree.Registry, hideAccesses bool) {
	registry.Password = ""
	registry.ManagementConfiguration = nil
	if hideAccesses {
		registry.RegistryAccesses = nil
	}
}
