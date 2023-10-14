package endpoints

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
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
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	var registries []portaineree.Registry
	err = handler.DataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
		registries, err = handler.listRegistries(tx, r, portainer.EndpointID(endpointID))
		return err
	})
	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, registries)
}

func (handler *Handler) listRegistries(tx dataservices.DataStoreTx, r *http.Request, endpointID portainer.EndpointID) ([]portaineree.Registry, error) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	user, err := tx.User().Read(securityContext.UserID)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve user from the database", err)
	}

	endpoint, err := tx.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if tx.IsErrObjectNotFound(err) {
		return nil, httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return nil, httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	isAdminOrEndpointAdmin, err := security.IsAdminOrEndpointAdmin(r, tx, endpoint.ID)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to check user role", err)
	}

	registries, err := tx.Registry().ReadAll()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve registries from the database", err)
	}

	registries, handleError := handler.filterRegistriesByAccess(r, registries, endpoint, user, securityContext.UserMemberships)
	if handleError != nil {
		return nil, handleError
	}

	for idx := range registries {
		hideRegistryFields(&registries[idx], !isAdminOrEndpointAdmin)
	}

	return registries, err
}

func (handler *Handler) filterRegistriesByAccess(r *http.Request, registries []portaineree.Registry, endpoint *portaineree.Endpoint, user *portaineree.User, memberships []portainer.TeamMembership) ([]portaineree.Registry, *httperror.HandlerError) {
	if !endpointutils.IsKubernetesEndpoint(endpoint) {
		return security.FilterRegistries(registries, user, memberships, endpoint.ID), nil
	}

	return handler.filterKubernetesEndpointRegistries(r, registries, endpoint, user, memberships)
}

func (handler *Handler) filterKubernetesEndpointRegistries(r *http.Request, registries []portaineree.Registry, endpoint *portaineree.Endpoint, user *portaineree.User, memberships []portainer.TeamMembership) ([]portaineree.Registry, *httperror.HandlerError) {
	namespaceParam, _ := request.RetrieveQueryParameter(r, "namespace", true)
	isAdminOrEndpointAdmin, err := security.IsAdminOrEndpointAdmin(r, handler.DataStore, endpoint.ID)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to check user role", err)
	}

	if namespaceParam != "" {
		authorized, err := handler.isNamespaceAuthorized(endpoint, namespaceParam, user.ID, memberships, isAdminOrEndpointAdmin)
		if err != nil {
			return nil, httperror.NotFound("Unable to check for namespace authorization", err)
		}
		if !authorized {
			return nil, httperror.Forbidden("User is not authorized to use namespace", errors.New("user is not authorized to use namespace"))
		}

		return filterRegistriesByNamespaces(registries, endpoint.ID, []string{namespaceParam}), nil
	}

	if isAdminOrEndpointAdmin {
		return registries, nil
	}

	return handler.filterKubernetesRegistriesByUserRole(r, registries, endpoint, user)
}

func (handler *Handler) isNamespaceAuthorized(endpoint *portaineree.Endpoint, namespace string, userId portainer.UserID, memberships []portainer.TeamMembership, isAdminOrEndpointAdmin bool) (bool, error) {
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

func filterRegistriesByNamespaces(registries []portaineree.Registry, endpointId portainer.EndpointID, namespaces []string) []portaineree.Registry {
	filteredRegistries := []portaineree.Registry{}

	for _, registry := range registries {
		if registryAccessPoliciesContainsNamespace(registry.RegistryAccesses[endpointId], namespaces) {
			filteredRegistries = append(filteredRegistries, registry)
		}
	}

	return filteredRegistries
}

func registryAccessPoliciesContainsNamespace(registryAccess portainer.RegistryAccessPolicies, namespaces []string) bool {
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
	if err != nil {
		if errors.Is(err, security.ErrAuthorizationRequired) {
			return nil, httperror.Forbidden("User is not authorized", err)
		}

		return nil, httperror.InternalServerError("Unable to retrieve info from request context", err)
	}

	userNamespaces, err := handler.userNamespaces(endpoint, user)
	if err != nil {
		return nil, httperror.InternalServerError("unable to retrieve user namespaces", err)
	}

	return filterRegistriesByNamespaces(registries, endpoint.ID, userNamespaces), nil
}

func (handler *Handler) userNamespaces(endpoint *portaineree.Endpoint, user *portaineree.User) ([]string, error) {
	kcl, err := handler.K8sClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return nil, err
	}

	namespaceAuthorizations, err := handler.AuthorizationService.GetNamespaceAuthorizations(handler.DataStore, int(user.ID), *endpoint, kcl)
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
