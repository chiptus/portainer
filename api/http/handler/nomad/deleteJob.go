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
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid Nomad job identifier route variable", Err: err}
	}

	namespace, err := request.RetrieveQueryParameter(r, "namespace", false)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid query parameter: namespace", Err: err}
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to establish communication with Nomad server", Err: err}
	}

	err = nomadClient.DeleteJob(jobID, namespace)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to delete Nomad job", Err: err}
	}

	return response.Empty(w)
}
