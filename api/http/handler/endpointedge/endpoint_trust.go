package endpointedge

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
)

// endpointTrust
// @summary Trust an edge device
// @description **Access policy**: admin
// @tags edge, endpoints
// @accept json
// @produce json
// @param id path string true "Environment(Endpoint) Id"
// @success 204
// @failure 500
// @failure 400
// @router /endpoints/{id}/edge/trust [post]
func (handler *Handler) endpointTrust(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	if !endpointutils.IsEdgeEndpoint(endpoint) {
		return httperror.BadRequest("Environment is not an edge environment", nil)
	}

	endpoint.UserTrusted = true

	err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to persist environment changes inside the database", err)
	}

	return response.Empty(w)
}
