package endpointedge

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer/pkg/featureflags"
)

// endpointTrust
// @summary Trust an edge device
// @description **Access policy**: admin
// @tags edge, endpoints
// @accept json
// @produce json
// @param id path int true "Environment(Endpoint) Id"
// @success 204
// @failure 500
// @failure 400
// @router /endpoints/{id}/edge/trust [post]
func (handler *Handler) endpointTrust(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		err = trustEndpoint(handler.DataStore, endpoint.ID)
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			return trustEndpoint(tx, endpoint.ID)
		})
	}

	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.Empty(w)
}

func trustEndpoint(tx dataservices.DataStoreTx, ID portaineree.EndpointID) error {
	endpoint, err := tx.Endpoint().Endpoint(ID)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environment from the database", err)
	}

	if !endpointutils.IsEdgeEndpoint(endpoint) {
		return httperror.BadRequest("Environment is not an edge environment", nil)
	}

	endpoint.UserTrusted = true

	err = tx.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to persist environment changes inside the database", err)
	}

	return nil
}
