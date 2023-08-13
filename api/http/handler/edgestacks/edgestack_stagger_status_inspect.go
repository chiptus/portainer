package edgestacks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

type edgeStackStaggerStatusResponse struct {
	Status string `json:"status"`
}

// @id EdgeStackStaggerStatusInspect
// @summary Inspect an EdgeStack's parallel update status
// @description **Access policy**: administrator
// @tags edge_stacks
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "EdgeStack Id"
// @success 200 {object} edgeStackStaggerStatusResponse
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id}/stagger/status [get]
func (handler *Handler) edgeStackStaggerStatusInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge stack identifier route variable", err)
	}

	_, err = handler.DataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(edgeStackID))
	if err != nil {
		return handler.handlerDBErr(err, "Unable to find an edge stack with the specified identifier inside the database")
	}

	resp := edgeStackStaggerStatusResponse{"idle"}
	if handler.staggerService.IsEdgeStackUpdating(portaineree.EdgeStackID(edgeStackID)) {
		resp.Status = "updating"
	}

	return response.JSON(w, resp)
}
