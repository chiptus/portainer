package dockersnapshot

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	_ "github.com/docker/docker/api/types"
	_ "github.com/docker/docker/api/types/mount"
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
// @param containerId path int true "Container identifier"
// @success 200 {object} portainer.DockerContainerSnapshot "Success"
// @failure 404 "Environment not found"
// @router /docker/{environmentId}/snapshot/containers/{containerId} [get]
func (handler *Handler) containerInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	containerId, err := request.RetrieveRouteVariableValue(r, "containerId")
	if err != nil {
		return httperror.BadRequest("Invalid container identifier route variable", err)
	}

	environmentId, _ := request.RetrieveNumericRouteVariableValue(r, "id")

	environmentSnapshot, err := handler.dataStore.Snapshot().Read(portaineree.EndpointID(environmentId))
	if err != nil {
		return httperror.NotFound("Unable to find a snapshot", err)
	}

	if environmentSnapshot == nil || environmentSnapshot.Docker == nil {
		return response.JSON(w, []string{})
	}

	containers := environmentSnapshot.Docker.SnapshotRaw.Containers

	for _, container := range containers {
		if container.ID == containerId {
			return response.JSON(w, container)
		}
	}

	return httperror.NotFound("Unable to find a container with the specified identifier inside the environment snapshot", nil)
}
