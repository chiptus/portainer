package kubernetes

import (
	"net/http"

	models "github.com/portainer/portainer-ee/api/http/models/kubernetes"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id GetClusterRoles
// @summary Get a list of cluster roles
// @description Get a list of cluster roles for the given kubernetes environment
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/cluster_roles [get]
func (handler *Handler) getClusterRoles(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	cli, handlerErr := handler.getProxyKubeClient(r)
	if handlerErr != nil {
		return handlerErr
	}

	clusterRoles, err := cli.GetClusterRoles()
	if err != nil {
		return httperror.InternalServerError(
			"Failed to fetch cluster roles",
			err,
		)
	}

	return response.JSON(w, clusterRoles)
}

// @id deleteClusterRoles
// @summary Delete the provided cluster roles
// @description Delete the provided cluster roles for the given Kubernetes environment
// @description **Access policy**: administrator
// @tags rbac_enabled
// @security ApiKeyAuth
// @security jwt
// @produce text/plain
// @param id path int true "Environment(Endpoint) identifier"
// @param payload body models.K8sClusterRoleDeleteRequests true "Cluster roles to delete"
// @success 200 "Success"
// @failure 500 "Server error"
// @router /kubernetes/{id}/cluster_roles/delete [POST]
func (handler *Handler) deleteClusterRoles(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload models.K8sClusterRoleDeleteRequests
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	cli, handlerErr := handler.getProxyKubeClient(r)
	if handlerErr != nil {
		return handlerErr
	}

	err = cli.DeleteClusterRoles(payload)
	if err != nil {
		return httperror.InternalServerError("Failed to delete cluster roles", err)
	}

	return nil
}
