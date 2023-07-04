package kaas

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/http/middlewares"
)

// @id version
// @summary Get the current cluster version.
// @description Get the current cluster version.
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
// @router /cloud/endpoints/{environmentid}/version [get]
func (handler *Handler) version(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	// And that the user has access to the environment
	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, false)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	credentialID := endpoint.CloudProvider.CredentialID
	credential, err := handler.dataStore.CloudCredential().Read(models.CloudCredentialID(credentialID))
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve SSH credential information", err)
	}

	var version string
	switch provider := endpoint.CloudProvider.Provider; provider {
	case portaineree.CloudProviderMicrok8s:
		version, err = handler.cloudInfoService.Microk8sVersion(credential, int(endpoint.ID))
		if err != nil {
			return httperror.InternalServerError("Failed upgrading cluster", err)
		}

	default:
		return httperror.BadRequest("bad request", fmt.Errorf("version request not implemented for %s", provider))
	}

	return response.JSON(w, version)
}
