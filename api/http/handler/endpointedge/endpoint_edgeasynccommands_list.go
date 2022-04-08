package endpointedge

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
)

type edgeAsyncCommandsResponse struct {
	Commands []portaineree.EdgeAsyncCommand `json:"commands"`
}

// @id EndpointEdgeAsyncCommands
// @summary Get environment(endpoint) commands
// @description environment(endpoint) for edge agent to check list of commands
// @description **Access policy**: restricted only to Edge environments(endpoints)
// @tags endpoints
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Environment(Endpoint) identifier"
// @success 200 {object} edgeAsyncCommandsResponse "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access environment(endpoint)"
// @failure 404 "Environment(Endpoint) not found"
// @failure 500 "Server error"
// @router /endpoints/{id}/edge/commands [get]
func (handler *Handler) endpointEdgeAsyncCommands(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.BadRequest("Unable to find an environment on request context", err)
	}

	if !endpointutils.IsEdgeEndpoint(endpoint) {
		return httperror.BadRequest("Invalid environment type", err)
	}

	commands, err := handler.DataStore.EdgeAsyncCommand().EndpointCommands(endpoint.ID)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve commands from the database", err)
	}

	commandsResponse := edgeAsyncCommandsResponse{
		Commands: commands,
	}

	return response.JSON(w, commandsResponse)
}
