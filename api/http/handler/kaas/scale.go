package kaas

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/providers"
	"github.com/portainer/portainer-ee/api/http/middlewares"
)

// @id addNodes
// @summary Add nodes to the cluster (scale up).
// @description Add control plane and worker nodes to the cluster.
// @description **Access policy**: authenticated
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param environmentid path int true "Environment(Endpoint) identifier"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/endpoints/{environmentid}/nodes/add [post]
func (handler *Handler) addNodes(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	var scalingRequest portaineree.CloudScalingRequest
	switch provider := endpoint.CloudProvider.Provider; provider {
	case portaineree.CloudProviderMicrok8s:
		var testssh bool
		err = request.RetrieveJSONQueryParameter(r, "testssh", &testssh, true)
		if err != nil {
			return httperror.BadRequest("Query parameter error", err)
		}

		if testssh {
			return handler.sshTestNodeIPs(w, r)
		}

		var p providers.Microk8sScaleClusterPayload
		err = request.DecodeAndValidateJSONPayload(r, &p)

		scalingRequest = &cloud.Microk8sScalingRequest{
			EndpointID:       endpoint.ID,
			MasterNodesToAdd: p.MasterNodesToAdd,
			WorkerNodesToAdd: p.WorkerNodesToAdd,
		}

	default:
		return httperror.BadRequest("Invalid request payload", fmt.Errorf("scaling from Portainer is not implemented for %s", provider))
	}

	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	handler.cloudManagementService.SubmitRequest(scalingRequest)
	return response.JSON(w, scalingRequest)
}

// @id removeNodes
// @summary Remove nodes from the cluster and uninstall MicroK8s from them.
// @description Remove nodes from the cluster and uninstall MicroK8s from them.
// @description **Access policy**: authenticated
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param environmentid path int true "Environment(Endpoint) identifier"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/endpoints/{environmentid}/nodes/remove [post]
func (handler *Handler) removeNodes(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	var scalingRequest portaineree.CloudScalingRequest
	switch provider := endpoint.CloudProvider.Provider; provider {
	case portaineree.CloudProviderMicrok8s:
		var p providers.Microk8sScaleClusterPayload
		err = request.DecodeAndValidateJSONPayload(r, &p)

		scalingRequest = &cloud.Microk8sScalingRequest{
			EndpointID:    endpoint.ID,
			NodesToRemove: p.NodesToRemove,
		}

	default:
		return httperror.BadRequest("Invalid request payload", fmt.Errorf("scaling from Portainer is not implemented for %s", provider))
	}

	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	handler.cloudManagementService.SubmitRequest(scalingRequest)
	return response.JSON(w, scalingRequest)
}