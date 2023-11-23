package endpoints

import (
	"errors"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/snapshot"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EndpointSnapshot
// @summary Snapshots an environment(endpoint)
// @description Snapshots an environment(endpoint)
// @description **Access policy**: authenticated
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Environment(Endpoint) identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id}/snapshot [post]
func (handler *Handler) endpointSnapshot(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid environment identifier route variable", err)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	if !snapshot.SupportDirectSnapshot(endpoint) && endpoint.Type != portaineree.EdgeAgentOnNomadEnvironment {
		return httperror.BadRequest("Snapshots not supported for this environment", errors.New("Snapshots not supported for this environment"))
	}

	snapshotError := handler.SnapshotService.SnapshotEndpoint(endpoint)

	latestEndpointReference, err := handler.DataStore.Endpoint().Endpoint(endpoint.ID)
	if latestEndpointReference == nil {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	}

	latestEndpointReference.Status = portainer.EndpointStatusUp
	if snapshotError != nil {
		latestEndpointReference.Status = portainer.EndpointStatusDown
	}

	latestEndpointReference.Agent.PreviousVersion = endpoint.Agent.PreviousVersion
	latestEndpointReference.Agent.Version = endpoint.Agent.Version

	err = handler.DataStore.Endpoint().UpdateEndpoint(latestEndpointReference.ID, latestEndpointReference)
	if err != nil {
		return httperror.InternalServerError("Unable to persist environment changes inside the database", err)
	}

	return response.Empty(w)
}
