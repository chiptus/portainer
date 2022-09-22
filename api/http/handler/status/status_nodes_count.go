package status

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	statusutil "github.com/portainer/portainer-ee/api/internal/nodes"
)

type nodesCountResponse struct {
	Nodes int `json:"nodes"`
}

// @id statusNodesCount
// @summary Retrieve the count of nodes
// @description **Access policy**: authenticated
// @security ApiKeyAuth
// @security jwt
// @tags status
// @produce json
// @success 200 {object} nodesCountResponse "Success"
// @failure 500 "Server error"
// @router /status/nodes [get]
func (handler *Handler) statusNodesCount(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	snapshots, err := handler.dataStore.Snapshot().Snapshots()
	if err != nil {
		return httperror.InternalServerError("Failed to get snapshot list", err)
	}

	nodes := statusutil.NodesCount(snapshots)
	return response.JSON(w, &nodesCountResponse{Nodes: nodes})
}
