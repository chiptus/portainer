package containers

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker/images"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	"github.com/rs/zerolog/log"

	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// containerImageStatus
// @id containerImageStatus
// @summary Fetch image status for container
// @description
// @description **Access policy**:
// @tags docker
// @security jwt
// @param environmentId path int true "Environment identifier"
// @param containerId path int true "Container identifier"
// @accept json
// @produce json
// @success 200 "Success"
// @failure 400 "Bad request"
// @failure 500 "Internal server error"
// @router /docker/{environmentId}/containers/{containerId}/image_status [get]
func (handler *Handler) containerImageStatus(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	containerID, err := request.RetrieveRouteVariableValue(r, "containerId")
	if err != nil {
		return httperror.BadRequest("Invalid containerId", err)
	}

	s, err := images.CachedResourceImageStatus(containerID)
	if err == nil {
		return response.JSON(w, &images.StatusResponse{Status: s, Message: ""})
	}

	log.Debug().Err(err).Str("containerId", containerID).Msg("No image status found from cache for container")

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.NotFound("Unable to find an environment on request context", err)
	}

	agentTargetHeader := r.Header.Get(portaineree.PortainerAgentTargetHeader)

	if err != nil {
		return httperror.InternalServerError("Unable to connect to the Docker daemon", err)
	}

	digestCli := images.NewClientWithRegistry(images.NewRegistryClient(handler.dataStore), handler.dockerClientFactory)

	s, err = digestCli.ContainerImageStatus(r.Context(), containerID, endpoint, agentTargetHeader)
	if err != nil {
		return httperror.InternalServerError("Unable to get the status of this image", err)
	}

	images.CacheResourceImageStatus(containerID, s)

	return response.JSON(w, &images.StatusResponse{Status: s, Message: ""})
}
