package users

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
)

// namespaceMapping is a struct created only for swagger API generation purposes
type namespaceMapping map[int]map[string]portainer.Authorizations

// @id UserNamespaces
// @summary Retrieves all k8s namespaces for an user
// @description Retrieves user's role authorizations of all namespaces in all k8s environments(endpoints)
// @description **Access policy**: restricted
// @tags users
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "User identifier"
// @success 200 {object} namespaceMapping "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "User not found"
// @failure 500 "Server error"
// @router /users/{id}/namespaces [get]
func (handler *Handler) userNamespaces(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	userID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid user identifier route variable", err}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user authentication token", err}
	}

	if tokenData.Role != portainer.AdministratorRole && tokenData.ID != portainer.UserID(userID) {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to retrieve user namespaces", errors.ErrUnauthorized}
	}

	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user namespace data", err}
	}

	// key: endpointID, value: a map between namespace and user's role authorizations
	results := make(map[int]map[string]portainer.Authorizations)
	for _, endpoint := range endpoints {

		// skip non k8s environments(endpoints)
		if endpoint.Type != portainer.KubernetesLocalEnvironment &&
			endpoint.Type != portainer.AgentOnKubernetesEnvironment &&
			endpoint.Type != portainer.EdgeAgentOnKubernetesEnvironment {
			continue
		}

		kcl, err := handler.K8sClientFactory.GetKubeClient(&endpoint)
		if err != nil {
			break
		}

		namespaceAuthorizations, err := handler.AuthorizationService.GetNamespaceAuthorizations(userID, endpoint, kcl)
		if err != nil {
			break
		}

		results[int(endpoint.ID)] = namespaceAuthorizations
	}

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user namespace data", err}
	}

	return response.JSON(w, results)
}
