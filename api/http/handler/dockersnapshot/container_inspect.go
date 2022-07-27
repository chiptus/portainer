package dockersnapshot

import (
	"net/http"

	_ "github.com/docker/docker/api/types"
	_ "github.com/docker/docker/api/types/mount"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/http/middlewares"
)

// @id snapshotContainerInspect
// @summary Fetch container from a snapshot
// @description
// @description **Access policy**:
// @tags endpoints,docker
// @security jwt
// @accept json
// @produce json
// @param environmentId path int true "Environment identifier"
// @param edgeStackId query int false "Edge stack identifier, will return only containers for this edge stack"
// @success 200 {object} []types.Container "Success"
// @failure 404 "Environment not found"
// @router /docker/{environmentId}/snapshot/container/{containerId} [get]
func (handler *Handler) containerInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	containerId, err := request.RetrieveRouteVariableValue(r, "containerId")
	if err != nil {
		return httperror.BadRequest("Invalid container identifier route variable", err)
	}

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.NotFound("Unable to find an environment on request context", err)
	}

	if len(endpoint.Snapshots) == 0 {
		return response.JSON(w, []string{})
	}

	snapshot := endpoint.Snapshots[0]
	containers := snapshot.SnapshotRaw.Containers

	for _, container := range containers {
		if container.ID == containerId {
			return response.JSON(w, container)
		}
	}

	return httperror.NotFound("Unable to find a container with the specified identifier inside the environment snapshot", nil)
}
