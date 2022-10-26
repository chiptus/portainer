package dockersnapshot

import (
	"fmt"
	"github.com/portainer/portainer-ee/api/docker/consts"
	"net/http"

	"github.com/docker/docker/api/types"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainerDsErrors "github.com/portainer/portainer/api/dataservices/errors"
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
// @success 200 {object} types.Container[] "Success"
// @failure 404 "Environment not found"
// @router /docker/{environmentId}/snapshot/containers [get]
func (handler *Handler) containersList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackId, _ := request.RetrieveNumericQueryParameter(r, "edgeStackId", true)
	environmentId, _ := request.RetrieveNumericRouteVariableValue(r, "id")

	environmentSnapshot, err := handler.dataStore.Snapshot().Snapshot(portaineree.EndpointID(environmentId))
	if err != nil {
		return response.JSON(w, []string{})
	}

	snapshot := environmentSnapshot.Docker
	containers := snapshot.SnapshotRaw.Containers

	if edgeStackId != 0 {
		edgeStack, err := handler.dataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(edgeStackId))

		if err != nil {
			statusCode := http.StatusNotFound
			if err != portainerDsErrors.ErrObjectNotFound {
				statusCode = http.StatusInternalServerError
			}
			return httperror.NewError(statusCode, "Unable to find an edge stack with the specified identifier inside the database", err)
		}

		containers = filterContainersByEdgeStack(containers, edgeStack)
	}

	return response.JSON(w, containers)
}

func filterContainersByEdgeStack(containers []types.Container, edgeStack *portaineree.EdgeStack) []types.Container {
	stackName := fmt.Sprintf("edge_%s", edgeStack.Name)
	filteredContainers := []types.Container{}

	for _, container := range containers {
		if container.Labels[consts.ComposeStackNameLabel] == stackName || container.Labels[consts.SwarmStackNameLabel] == stackName {
			filteredContainers = append(filteredContainers, container)
		}
	}

	return filteredContainers
}
