package nomad

import (
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
	"net/http"
)

type (
	LeaderPayload struct {
		Leader string `json:"Leader"`
	}
)

// @id getLeader
// @summary returns the address of the current leader in the region
// @description **Access policy**: authenticated users
// @tags nomad
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 "Success"
// @failure 404 "Endpoint not found"
// @failure 500 "Server error"
// @router /nomad/status [get]
func (handler *Handler) getLeader(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid query parameter: endpointId", Err: err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
	if err == portainerDsErrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find an environment with the specified identifier inside the database", Err: err}
	}

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to establish communication with Nomad server", Err: err}
	}

	leader, err := nomadClient.Leader()
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to get leader", Err: err}
	}

	leaderPayload := LeaderPayload{
		Leader: leader,
	}

	return response.JSON(w, leaderPayload)
}
