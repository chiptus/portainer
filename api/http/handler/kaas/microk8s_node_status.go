package kaas

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/cloud/microk8s"
	"github.com/portainer/portainer-ee/api/database/models"
)

// @id microk8sGetNodeStatus
// @summary Get the MicroK8s status for a control plane node.
// @description Get the MicroK8s status for a control plane node in a MicroK8s cluster.
// @description **Access policy**: authenticated
// @tags kaas
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param nodeIP query string true "The external node ip of the control plane node."
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 500 "Server error"
// @failure 503 "Missing configuration"
// @router /cloud/endpoints/{environmentid}/nodes/nodestatus [get]
func (handler *Handler) microk8sGetNodeStatus(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	environmentId, err := request.RetrieveNumericRouteVariableValue(r, "environmentid")

	nodeIP, err := request.RetrieveQueryParameter(r, "nodeIP", false)
	if err != nil {
		return httperror.BadRequest("The nodeIP query parameter must be defined", err)
	}
	if nodeIP == "" {
		return httperror.BadRequest("The nodeIP query parameter must have a value", err)
	}

	endpoint, err := handler.dataStore.Endpoint().Endpoint(portaineree.EndpointID(environmentId))
	if handler.dataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find the environment in the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find the environment in the database", err)
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint, true)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	credentialId := endpoint.CloudProvider.CredentialID
	credential, err := handler.dataStore.CloudCredential().Read(models.CloudCredentialID(credentialId))
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve SSH credential information", err)
	}
	status, err := handler.cloudInfoService.Microk8sGetStatus(credential, int(endpoint.ID), nodeIP)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve MicroK8s status", err)
	}
	nodeStatusResponse := microk8s.Microk8sNodeStatusResponse{
		Status: status,
	}

	return response.JSON(w, nodeStatusResponse)
}
