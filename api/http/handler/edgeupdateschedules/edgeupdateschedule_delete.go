package edgeupdateschedules

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"

	"github.com/portainer/portainer-ee/api/http/middlewares"
)

// @id EdgeUpdateScheduleDelete
// @summary Deletes an Edge Update Schedule
// @description **Access policy**: administrator
// @tags edge_update_schedules
// @security ApiKeyAuth
// @security jwt
// @success 204
// @failure 500
// @router /edge_update_schedules/{id} [delete]
func (handler *Handler) delete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	item, err := middlewares.FetchItem[edgetypes.UpdateSchedule](r, contextKey)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	edgeStack, err := handler.dataStore.EdgeStack().EdgeStack(item.EdgeStackID)
	if err != nil {
		return httperror.InternalServerError("failed fetching edge stack", err)
	}

	err = handler.edgeStacksService.DeleteEdgeStack(edgeStack.ID, edgeStack.EdgeGroups)
	if err != nil {
		return httperror.InternalServerError("failed deleting schedule edge stack", err)
	}

	err = handler.updateService.DeleteSchedule(item.ID)
	if err != nil {
		return httperror.InternalServerError("Unable to delete the edge update schedule", err)
	}

	return response.Empty(w)
}
