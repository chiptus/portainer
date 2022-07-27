package dockersnapshot

import (
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/docker"
	"github.com/portainer/portainer-ee/api/http/middlewares"
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

	endpoint, err := middlewares.FetchEndpoint(r)
	if err != nil {
		return httperror.NotFound("Unable to find an environment on request context", err)
	}

	if len(endpoint.Snapshots) == 0 {
		return response.JSON(w, []string{})
	}

	snapshot := endpoint.Snapshots[0]
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
		if container.Labels[docker.ComposeStackNameLabel] == stackName || container.Labels[docker.SwarmStackNameLabel] == stackName {
			filteredContainers = append(filteredContainers, container)
		}
	}

	return filteredContainers
}
