package edgeupdateschedules

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
)

type inspectResponse struct {
	*edgetypes.UpdateSchedule
	EdgeGroupIds  []portaineree.EdgeGroupID `json:"edgeGroupIds"`
	ScheduledTime string                    `json:"scheduledTime"`
	IsActive      bool                      `json:"isActive"`
}

// @id EdgeUpdateScheduleInspect
// @summary Returns the Edge Update Schedule with the given ID
// @description **Access policy**: administrator
// @tags edge_update_schedules
// @security ApiKeyAuth
// @security jwt
// @param id path int true "EdgeUpdate Id"
// @produce json
// @success 200 {object} decoratedUpdateSchedule
// @failure 500
// @router /edge_update_schedules/{id} [get]
func (handler *Handler) inspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	item, err := middlewares.FetchItem[edgetypes.UpdateSchedule](r, contextKey)
	if err != nil {
		return httperror.InternalServerError(err.Error(), err)
	}

	includeEdgeStack, _ := request.RetrieveBooleanQueryParameter(r, "includeEdgeStack", true)
	if !includeEdgeStack {
		return response.JSON(w, item)
	}

	edgeStack, err := handler.dataStore.EdgeStack().EdgeStack(item.EdgeStackID)
	if err != nil {
		return httperror.InternalServerError("unable to get edge stack", err)
	}

	isActive := false
	for _, envStatus := range edgeStack.Status {
		if !envStatus.Details.Pending {
			isActive = true
			break
		}
	}

	decoratedItem := &inspectResponse{
		UpdateSchedule: item,
		EdgeGroupIds:   edgeStack.EdgeGroups,
		IsActive:       isActive,
		ScheduledTime:  edgeStack.ScheduledTime,
	}

	if err != nil {
		return httperror.InternalServerError("Unable to decorate the edge update schedule", err)
	}

	return response.JSON(w, decoratedItem)
}
