package kaas

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/http/middlewares"
)

// @id upgrade
// @summary Upgrade the cluster to the next stable kubernetes version.
// @description Upgrade the cluster to the next stable kubernetes version.
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
// @router /cloud/endpoints/{environmentid}/upgrade [post]
func (handler *Handler) upgrade(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	if endpoint.CloudProvider == nil {
		return httperror.BadRequest("this is not a cloud environment", err)
	}

	// And that the user has access to the environment
	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, false)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	var upgradeRequest portaineree.CloudUpgradeRequest
	switch provider := endpoint.CloudProvider.Provider; provider {
	case portaineree.CloudProviderMicrok8s:
		upgradeRequest = &cloud.Microk8sUpgradeRequest{
			EndpointID: endpoint.ID,
		}
	default:
		return httperror.BadRequest("Invalid request payload", fmt.Errorf("upgrade not allowed for provider: %s", endpoint.CloudProvider.Provider))
	}

	handler.cloudManagementService.SubmitRequest(upgradeRequest)
	return response.JSON(w, upgradeRequest)
}
