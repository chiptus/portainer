package kubernetes

import (
	"net/http"

	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id GetRolesBindings
// @summary Get a list of role bindings
// @description Get a list of role bindings for the given kubernetes environment in the given namespace
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @param namespace path string true "Kubernetes namespace"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/namespaces/{namespace}/roles [get]
func (h *Handler) getRoleBindings(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	namespace, err := request.RetrieveRouteVariableValue(r, "namespace")
	if err != nil {
		return httperror.BadRequest("Invalid namespace identifier route variable", err)
	}

	cli, handlerErr := h.getProxyKubeClient(r)
	if handlerErr != nil {
		return handlerErr
	}

	clusterRoleBindings, err := cli.GetRoleBindings(namespace)
	if err != nil {
		return httperror.InternalServerError("Failed to fetch role bindings", err)
	}

	return response.JSON(w, clusterRoleBindings)
}

// @id DeleteRoleBindings
// @summary Delete the provided role bindings
// @description Delete the provided role bindings for the given Kubernetes environment
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @param payload body models.K8sRoleDeleteRequests true "Role bindings to delete"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/role_bindings/delete [POST]
func (h *Handler) deleteRoleBindings(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload models.K8sRoleBindingDeleteRequests
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	cli, handlerErr := h.getProxyKubeClient(r)
	if err != nil {
		return handlerErr
	}

	err = cli.DeleteRoleBindings(payload)
	if err != nil {
		return httperror.InternalServerError("Failed to delete role bindings", err)
	}

	return nil
}
