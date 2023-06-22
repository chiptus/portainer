package kaas

import (
	"fmt"
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/database/models"
)

// @id microk8sAddons
// @summary Get a list of addons which are installed in a MicroK8s cluster.
// @description The information returned can be used to query the MircoK8s cluster.
// @description **Access policy**: authenticated
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param environmentID query int true "The environment ID of the cluster within Portainer."
// @param credentialID query int true "The credential ID to use to connect to a node in the MicroK8s cluster."
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/microk8s/addons [get]
func (handler *Handler) microk8sAddons(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	environmentID, err := request.RetrieveQueryParameter(r, "environmentID", true)
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier", err)
	}

	// Check if the environment exists
	formattedEnvironmentID, err := strconv.Atoi(environmentID)
	if err != nil {
		return httperror.InternalServerError("Failed to parse environmentID", err)
	}
	environment, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(formattedEnvironmentID))
	if handler.dataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	// And that the user has access to the environment
	err = handler.requestBouncer.AuthorizedEndpointOperation(r, environment, false)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	credentialId, _ := request.RetrieveNumericQueryParameter(r, "credentialID", true)
	if credentialId == 0 {
		return httperror.InternalServerError("Missing credential id in the query parameter", err)
	}

	credential, err := handler.dataStore.CloudCredential().Read(models.CloudCredentialID(credentialId))
	if err != nil {
		return httperror.InternalServerError(fmt.Sprintf("Unable to retrieve SSH credential information"), err)
	}

	microK8sInfo, err := handler.cloudClusterInfoService.Microk8sGetAddons(credential, formattedEnvironmentID)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve MicroK8s information", err)
	}

	return response.JSON(w, microK8sInfo)
}
