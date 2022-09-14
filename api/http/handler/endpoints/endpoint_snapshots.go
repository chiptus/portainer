package endpoints

import (
	"log"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
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
	endpoints, err := handler.dataStore.Endpoint().Endpoints()
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

		latestEndpointReference, err := handler.dataStore.Endpoint().Endpoint(endpoint.ID)
		if latestEndpointReference == nil {
			log.Printf("background schedule error (environment snapshot). Environment not found inside the database anymore (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, err)
			continue
		}

		endpoint.Status = portaineree.EndpointStatusUp
		if snapshotError != nil {
			log.Printf("background schedule error (environment snapshot). Unable to create snapshot (endpoint=%s, URL=%s) (err=%s)\n", endpoint.Name, endpoint.URL, snapshotError)
			endpoint.Status = portaineree.EndpointStatusDown
		}

		latestEndpointReference.Snapshots = endpoint.Snapshots
		latestEndpointReference.Kubernetes.Snapshots = endpoint.Kubernetes.Snapshots
		latestEndpointReference.Nomad.Snapshots = endpoint.Nomad.Snapshots
		latestEndpointReference.Agent.Version = endpoint.Agent.Version

		err = handler.dataStore.Endpoint().UpdateEndpoint(latestEndpointReference.ID, latestEndpointReference)
		if err != nil {
			return httperror.InternalServerError("Unable to persist environment changes inside the database", err)
		}
	}

	return response.Empty(w)
}
