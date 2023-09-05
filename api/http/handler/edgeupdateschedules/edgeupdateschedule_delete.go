package edgeupdateschedules

import (
	"net/http"

	"github.com/portainer/portainer-ee/api/http/middlewares"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeUpdateScheduleDelete
// @summary Deletes an Edge Update Schedule
// @description **Access policy**: administrator
// @tags edge_update_schedules
// @security ApiKeyAuth
// @param id path int true "EdgeUpdate Id"
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

	err = handler.edgeStacksService.DeleteEdgeStack(handler.dataStore, edgeStack.ID, edgeStack.EdgeGroups)
	if err != nil {
		return httperror.InternalServerError("failed deleting schedule edge stack", err)
	}

	if len(edgeStack.EdgeGroups) > 0 {
		err = handler.dataStore.EdgeGroup().Delete(edgeStack.EdgeGroups[0])
		if err != nil {
			return httperror.InternalServerError("failed deleting schedule edge group", err)
		}
	}

	err = handler.updateService.DeleteSchedule(item.ID)
	if err != nil {
		return httperror.InternalServerError("Unable to delete the edge update schedule", err)
	}

	return response.Empty(w)
}
