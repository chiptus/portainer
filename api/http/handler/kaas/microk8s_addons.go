package kaas

import (
	"fmt"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud"
	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/http/handler/kaas/providers"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id microk8sGetAddons
// @summary Get a list of addons which are installed in a MicroK8s cluster.
// @description The information returned can be used to query the MircoK8s cluster.
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
// @router /cloud/endpoints/{environmentid}/addons [get]
func (handler *Handler) microk8sGetAddons(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	// And that the user has access to the environment
	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, false)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	if endpoint.CloudProvider == nil {
		return httperror.BadRequest("bad request", fmt.Errorf("this is not a cloud environment"))
	}

	if endpoint.CloudProvider.Provider != portaineree.CloudProviderMicrok8s {
		return httperror.BadRequest("bad request", fmt.Errorf("this cluster was not provisioned by Portainer"))
	}

	credentialId := endpoint.CloudProvider.CredentialID
	credential, err := handler.dataStore.CloudCredential().Read(models.CloudCredentialID(credentialId))
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve SSH credential information", err)
	}

	microK8sInfo, err := handler.cloudInfoService.Microk8sGetAddons(credential, int(endpoint.ID))
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve MicroK8s information", err)
	}

	return response.JSON(w, microK8sInfo)
}

// @id microk8sUpdateAddons
// @summary Get a list of addons which are installed in a MicroK8s cluster.
// @description The information returned can be used to query the MircoK8s cluster.
// @description **Access policy**: authenticated
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param environmentid path int true "Environment(Endpoint) identifier"
// @param addons body providers.Microk8sUpdateAddonsPayload true "The list of addons to install in the cluster."
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/endpoints/{environmentid}/addons [post]
func (handler *Handler) microk8sUpdateAddons(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	// And that the user has access to the environment
	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, false)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	if endpoint.CloudProvider == nil {
		return httperror.BadRequest("bad request", fmt.Errorf("this is not a cloud environment"))
	}

	if endpoint.CloudProvider.Provider != portaineree.CloudProviderMicrok8s {
		return httperror.BadRequest("bad request", fmt.Errorf("this cluster was not provisioned by Portainer"))
	}

	var p providers.Microk8sUpdateAddonsPayload
	err = request.DecodeAndValidateJSONPayload(r, &p)
	if err != nil {
		return httperror.BadRequest("Invalid addons request payload", err)
	}

	updateAddonsRequest := &cloud.Microk8sUpdateAddonsRequest{
		EndpointID: endpoint.ID,
		Addons:     p.Addons,
	}

	handler.cloudManagementService.SubmitRequest(updateAddonsRequest)

	// update the environment status to processing before returning
	endpoint.StatusMessage = portaineree.EndpointStatusMessage{
		Summary:         "Updating addons",
		Detail:          "Enabling/Disabling MicroK8s addons",
		Operation:       "addons",
		OperationStatus: "processing",
	}
	err = handler.dataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to update the environment", err)
	}

	return response.JSON(w, updateAddonsRequest)
}
