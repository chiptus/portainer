package nomad

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/http/middlewares"
)

// @id deleteJob
// @summary Delete a job
// @description Job ID and namespace params are required
// @description **Access policy**: administrator
// @tags nomad
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 "Success"
// @failure 500 "Server error"
// @router /nomad/endpoints/{endpointID}/jobs/{id} [delete]
func (handler *Handler) deleteJob(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	jobID, err := request.RetrieveRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Nomad job identifier route variable", err)
	}

	namespace, err := request.RetrieveQueryParameter(r, "namespace", false)
	if err != nil {
		return httperror.BadRequest("Invalid query parameter: namespace", err)
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to establish communication with Nomad server", err)
	}

	err = nomadClient.DeleteJob(jobID, namespace)
	if err != nil {
		return httperror.InternalServerError("Unable to delete Nomad job", err)
	}

	return response.Empty(w)
}
