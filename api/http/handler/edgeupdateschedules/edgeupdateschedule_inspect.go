package edgeupdateschedules

import (
	"net/http"
	"slices"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/http/middlewares"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
	"github.com/rs/zerolog/log"
)

type inspectResponse struct {
	*edgetypes.UpdateSchedule
	ScheduledTime  string                 `json:"scheduledTime"`
	IsActive       bool                   `json:"isActive"`
	EnvironmentIDs []portainer.EndpointID `json:"environmentIds"`
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

	if len(edgeStack.EdgeGroups) != 1 {
		if len(edgeStack.EdgeGroups) == 0 {
			return httperror.BadRequest("edge stack has no edge groups", nil)
		}

		log.Warn().
			Int("edgeStackID", int(edgeStack.ID)).
			Int("edgeUpdateID", int(item.ID)).
			Int("edgeGroupCount", len(edgeStack.EdgeGroups)).
			Msg("edge stack has multiple edge groups")
	}

	edgeGroup, err := handler.dataStore.EdgeGroup().Read(edgeStack.EdgeGroups[0])
	if err != nil {
		return httperror.InternalServerError("unable to get edge group", err)
	}

	isActive := isUpdateActive(edgeStack)

	decoratedItem := &inspectResponse{
		UpdateSchedule: item,
		IsActive:       isActive,
		ScheduledTime:  edgeStack.ScheduledTime,
		// the edge update group is always static
		EnvironmentIDs: edgeGroup.Endpoints,
	}

	if err != nil {
		return httperror.InternalServerError("Unable to decorate the edge update schedule", err)
	}

	return response.JSON(w, decoratedItem)
}

func isUpdateActive(edgeStack *portaineree.EdgeStack) bool {
	for _, envStatus := range edgeStack.Status {
		if slices.ContainsFunc(envStatus.Status, func(s portainer.EdgeStackDeploymentStatus) bool {
			return s.Type != portainer.EdgeStackStatusPending
		}) {

			return true
		}
	}

	return false
}
