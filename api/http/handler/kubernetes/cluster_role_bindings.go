package kubernetes

import (
	"net/http"

	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id GetClusterRoleBindings
// @summary Get a list of cluster role bindings
// @description Get a list of all cluster role bindings for the given kubernetes environment
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/cluster_role_bindings [get]
func (handler *Handler) getClusterRoleBindings(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	cli, handlerErr := handler.getProxyKubeClient(r)
	if handlerErr != nil {
		return handlerErr
	}

	clusterRoleBindings, err := cli.GetClusterRoleBindings()
	if err != nil {
		return httperror.InternalServerError("Failed to fetch cluster roles", err)
	}

	return response.JSON(w, clusterRoleBindings)
}

// @id DeleteClusterRoleBindings
// @summary Delete the provided cluster role bindings
// @description Delete the provided cluster role bindings for the given Kubernetes environment
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @param payload body models.K8sClusterRoleBindingDeleteRequests true "Cluster role bindings to delete"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/cluster_role_bindings/delete [POST]
func (handler *Handler) deleteClusterRoleBindings(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload models.K8sClusterRoleBindingDeleteRequests
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	cli, handlerErr := handler.getProxyKubeClient(r)
	if handlerErr != nil {
		return handlerErr
	}

	err = cli.DeleteClusterRoleBindings(payload)
	if err != nil {
		return httperror.InternalServerError("Failed to delete cluster role bindings", err)
	}

	return nil
}
