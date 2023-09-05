package nomad

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/http/middlewares"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id deleteJob
// @summary Delete a job
// @description Job ID and namespace params are required
// @description **Access policy**: administrator
// @tags nomad
// @security ApiKeyAuth
// @security jwt
// @param environmentId path int true "Environment identifier"
// @param id path int true "Job identifier"
// @produce json
// @success 200 "Success"
// @failure 500 "Server error"
// @router /nomad/endpoints/{environmentId}/jobs/{id} [delete]
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
