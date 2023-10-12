package kubernetes

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id GetKubeRoles
// @summary Get a list of roles
// @description Get a list of roles for the given kubernetes environment in the given namespace
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @param namespace path string true "Kubernetes namespace"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/namespaces/{namespace/roles [get]
func (h *Handler) getRoles(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest(
			"Invalid environment identifier route variable",
			err,
		)
	}

	namespace, err := request.RetrieveRouteVariableValue(r, "namespace")
	if err != nil {
		return httperror.BadRequest(
			"Invalid namespace identifier route variable",
			err,
		)
	}

	endpoint, err := h.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if h.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound(
			"Unable to find an environment with the specified identifier inside the database",
			err,
		)
	} else if err != nil {
		return httperror.InternalServerError(
			"Unable to find an environment with the specified identifier inside the database",
			err,
		)
	}

	cli, err := h.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return httperror.InternalServerError(
			"Unable to create Kubernetes client",
			err,
		)
	}

	roles, err := cli.GetRoles(namespace)
	if err != nil {
		return httperror.InternalServerError(
			"Failed to fetch roles",
			err,
		)
	}

	return response.JSON(w, roles)
}

// @id DeleteRoles
// @summary Delete the provided roles
// @description Delete the provided roles for the given Kubernetes environment
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @param payload body models.K8sRoleDeleteRequests true "Roles to delete "
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/roles/delete [POST]
func (h *Handler) deleteRoles(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest(
			"Invalid environment identifier route variable",
			err,
		)
	}

	var payload models.K8sRoleDeleteRequests
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest(
			"Invalid request payload",
			err,
		)
	}

	endpoint, err := h.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if h.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound(
			"Unable to find an environment with the specified identifier inside the database",
			err,
		)
	} else if err != nil {
		return httperror.InternalServerError(
			"Unable to find an environment with the specified identifier inside the database",
			err,
		)
	}

	cli, err := h.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return httperror.InternalServerError(
			"Unable to create Kubernetes client",
			err,
		)
	}

	err = cli.DeleteRoles(payload)
	if err != nil {
		return httperror.InternalServerError(
			"Failed to delete roles",
			err,
		)
	}

	return nil
}
