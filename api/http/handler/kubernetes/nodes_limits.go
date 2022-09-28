package kubernetes

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id getKubernetesNodesLimits
// @summary Get CPU and memory limits of all nodes within k8s cluster
// @description Get CPU and memory limits of all nodes within k8s cluster
// @description **Access policy**: authenticated
// @tags kubernetes
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {object} portaineree.K8sNodesLimits "Success"
// @failure 400 "Invalid request"
// @failure 401 "Unauthorized"
// @failure 403 "Permission denied"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /kubernetes/{id}/nodes_limits [get]
func (handler *Handler) getKubernetesNodesLimits(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	cli := handler.KubernetesClient
	nodesLimits, err := cli.GetNodesLimits()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve nodes limits", err)
	}

	return response.JSON(w, nodesLimits)
}
