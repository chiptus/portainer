package dockersnapshot

import (
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

type remoteManager struct {
	NodeID string
	Addr   string
}

type snapshotInfoSwarm struct {
	NodeID           string          `json:"NodeID"`
	NodeAddr         string          `json:"NodeAddr"`
	LocalNodeState   string          `json:"LocalNodeState"`
	ControlAvailable bool            `json:"ControlAvailable"`
	Error            string          `json:"Error"`
	RemoteManagers   []remoteManager `json:"RemoteManagers"`
}

type snapshotInfo struct {
	Containers        int
	ContainersRunning int
	ContainersPaused  int
	ContainersStopped int
	Images            int
	SystemTime        string
	Swarm             snapshotInfoSwarm
}

type snapshotResponse struct {
	Containers []portainer.DockerContainerSnapshot
	Volumes    volume.VolumeListOKBody
	Images     []types.ImageSummary
	Info       snapshotInfo
}

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
	if err != nil || environmentSnapshot == nil || environmentSnapshot.Docker == nil {
		return httperror.NotFound("Unable to find a snapshot", err)
	}

	snapshotResponse := snapshotResponse{
		Containers: environmentSnapshot.Docker.SnapshotRaw.Containers,
		Volumes:    environmentSnapshot.Docker.SnapshotRaw.Volumes,
		Images:     environmentSnapshot.Docker.SnapshotRaw.Images,
		Info: snapshotInfo{
			SystemTime:        environmentSnapshot.Docker.SnapshotRaw.Info.SystemTime,
			Containers:        environmentSnapshot.Docker.SnapshotRaw.Info.Containers,
			ContainersRunning: environmentSnapshot.Docker.SnapshotRaw.Info.ContainersRunning,
			ContainersPaused:  environmentSnapshot.Docker.SnapshotRaw.Info.ContainersPaused,
			ContainersStopped: environmentSnapshot.Docker.SnapshotRaw.Info.ContainersStopped,
			Images:            environmentSnapshot.Docker.SnapshotRaw.Info.Images,
			Swarm: snapshotInfoSwarm{
				NodeID: environmentSnapshot.Docker.SnapshotRaw.Info.Swarm.NodeID,
			},
		},
	}

	return response.JSON(w, snapshotResponse)
}
