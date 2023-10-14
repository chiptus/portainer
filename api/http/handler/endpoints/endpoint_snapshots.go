package endpoints

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/internal/snapshot"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/rs/zerolog/log"
)

// @id EndpointSnapshots
// @summary Snapshot all environment(endpoint)
// @description Snapshot all environments(endpoints)
// @description **Access policy**: administrator
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @success 204 "Success"
// @failure 500 "Server Error"
// @router /endpoints/snapshot [post]
func (handler *Handler) endpointSnapshots(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environments from the database", err)
	}

	for _, endpoint := range endpoints {
		if !snapshot.SupportDirectSnapshot(&endpoint) {
			continue
		}

		if endpoint.URL == "" {
			continue
		}

		snapshotError := handler.SnapshotService.SnapshotEndpoint(&endpoint)

		latestEndpointReference, err := handler.DataStore.Endpoint().Endpoint(endpoint.ID)
		if latestEndpointReference == nil {
			log.Debug().
				Str("endpoint", endpoint.Name).
				Str("URL", endpoint.URL).
				Err(err).
				Msg("background schedule error (environment snapshot), environment not found inside the database anymore")

			continue
		}

		endpoint.Status = portainer.EndpointStatusUp
		if snapshotError != nil {
			log.Debug().
				Str("endpoint", endpoint.Name).
				Str("URL", endpoint.URL).
				Err(snapshotError).
				Msg("background schedule error (environment snapshot), unable to create snapshot")

			endpoint.Status = portainer.EndpointStatusDown
		}

		latestEndpointReference.Agent.Version = endpoint.Agent.Version

		err = handler.DataStore.Endpoint().UpdateEndpoint(latestEndpointReference.ID, latestEndpointReference)
		if err != nil {
			return httperror.InternalServerError("Unable to persist environment changes inside the database", err)
		}
	}

	return response.Empty(w)
}
