package nomad

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/http/middlewares"
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
// @router /nomad/endpoints/{endpointID}/status [get]
func (handler *Handler) getLeader(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	nomadClient, err := handler.nomadClientFactory.GetClient(endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to establish communication with Nomad server", err)
	}

	leader, err := nomadClient.Leader()
	if err != nil {
		return httperror.InternalServerError("Unable to get leader", err)
	}

	leaderPayload := LeaderPayload{
		Leader: leader,
	}

	return response.JSON(w, leaderPayload)
}
