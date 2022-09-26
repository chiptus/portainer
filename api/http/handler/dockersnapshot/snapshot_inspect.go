package dockersnapshot

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

// @id snapshotInspect
// @summary Fetch latest snapshot of environment
// @description
// @description **Access policy**:
// @tags endpoints,docker
// @security jwt
// @accept json
// @produce json
// @param environmentId path int true "Environment identifier"
// @success 200 {object} portainer.DockerSnapshotRaw "Success"
// @failure 404 "Environment not found"
// @router /docker/{environmentId}/snapshot [get]
func (handler *Handler) snapshotInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	environmentId, _ := request.RetrieveNumericRouteVariableValue(r, "id")

	environmentSnapshot, err := handler.dataStore.Snapshot().Snapshot(portaineree.EndpointID(environmentId))
	if err != nil || environmentSnapshot == nil {
		return httperror.NotFound("Unable to find a snapshot", err)
	}

	snapshot := environmentSnapshot.Docker.SnapshotRaw

	return response.JSON(w, snapshot)
}
