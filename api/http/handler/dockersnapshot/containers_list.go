package dockersnapshot

import (
	"fmt"
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker/consts"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id snapshotContainersList
// @summary Fetch containers list from a snapshot
// @description
// @description **Access policy**:
// @tags endpoints,docker
// @security jwt
// @accept json
// @produce json
// @param environmentId path int true "Environment identifier"
// @param edgeStackId query int false "Edge stack identifier, will return only containers for this edge stack"
// @success 200 {object} portainer.DockerContainerSnapshot[] "Success"
// @failure 404 "Environment not found"
// @router /docker/{environmentId}/snapshot/containers [get]
func (handler *Handler) containersList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackId, _ := request.RetrieveNumericQueryParameter(r, "edgeStackId", true)
	environmentId, _ := request.RetrieveNumericRouteVariableValue(r, "id")

	environmentSnapshot, err := handler.dataStore.Snapshot().Read(portaineree.EndpointID(environmentId))
	if err != nil || environmentSnapshot == nil || environmentSnapshot.Docker == nil {
		return response.JSON(w, []string{})
	}

	containers := environmentSnapshot.Docker.SnapshotRaw.Containers

	if edgeStackId != 0 {
		edgeStack, err := handler.dataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(edgeStackId))

		if err != nil {
			statusCode := http.StatusNotFound
			if !handler.dataStore.IsErrObjectNotFound(err) {
				statusCode = http.StatusInternalServerError
			}
			return httperror.NewError(statusCode, "Unable to find an edge stack with the specified identifier inside the database", err)
		}

		containers = filterContainersByEdgeStack(containers, edgeStack)
	}

	return response.JSON(w, containers)
}

func filterContainersByEdgeStack(containers []portainer.DockerContainerSnapshot, edgeStack *portaineree.EdgeStack) []portainer.DockerContainerSnapshot {
	stackName := fmt.Sprintf("edge_%s", edgeStack.Name)
	filteredContainers := []portainer.DockerContainerSnapshot{}

	for _, container := range containers {
		if container.Labels[consts.ComposeStackNameLabel] == stackName || container.Labels[consts.SwarmStackNameLabel] == stackName {
			filteredContainers = append(filteredContainers, container)
		}
	}

	return filteredContainers
}
