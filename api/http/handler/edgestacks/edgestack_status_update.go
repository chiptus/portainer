package edgestacks

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	"github.com/rs/zerolog/log"
)

type updateStatusPayload struct {
	Error      string
	Status     *portaineree.EdgeStackStatusType
	EndpointID portaineree.EndpointID
}

func (payload *updateStatusPayload) Validate(r *http.Request) error {
	if payload.Status == nil {
		return errors.New("Invalid status")
	}

	if payload.EndpointID == 0 {
		return errors.New("Invalid EnvironmentID")
	}

	if *payload.Status == portaineree.StatusError && govalidator.IsNull(payload.Error) {
		return errors.New("Error message is mandatory when status is error")
	}

	return nil
}

// @id EdgeStackStatusUpdate
// @summary Update an EdgeStack status
// @description Authorized only if the request is done by an Edge Environment(Endpoint)
// @tags edge_stacks
// @accept json
// @produce json
// @param id path string true "EdgeStack Id"
// @success 200 {object} portaineree.EdgeStack
// @failure 500
// @failure 400
// @failure 404
// @failure 403
// @router /edge_stacks/{id}/status [put]
func (handler *Handler) edgeStackStatusUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid stack identifier route variable", err)
	}

	stack, err := handler.DataStore.EdgeStack().EdgeStack(portaineree.EdgeStackID(stackID))
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a stack with the specified identifier inside the database", err)
	}

	var payload updateStatusPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	// if the stack represents a successful remote update - skip it
	if endpointStatus, ok := stack.Status[payload.EndpointID]; ok && endpointStatus.Type == portaineree.EdgeStackStatusRemoteUpdateSuccess {
		return response.JSON(w, stack)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(payload.EndpointID)
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an environment with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an environment with the specified identifier inside the database", err)
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return httperror.Forbidden("Permission denied to access environment", err)
	}

	status := *payload.Status

	if stack.EdgeUpdateID != 0 {
		if status == portaineree.StatusError {
			err := handler.edgeUpdateService.RemoveActiveSchedule(payload.EndpointID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
			if err != nil {
				log.Warn().
					Err(err).
					Msg("Failed to remove active schedule")
			}
		}

		if status == portaineree.StatusOk {
			handler.edgeUpdateService.EdgeStackDeployed(endpoint.ID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
		}
	}

	err = handler.DataStore.EdgeStack().UpdateEdgeStackFunc(portaineree.EdgeStackID(stackID), func(edgeStack *portaineree.EdgeStack) {
		edgeStack.Status[payload.EndpointID] = portaineree.EdgeStackStatus{
			Type:       status,
			Error:      payload.Error,
			EndpointID: payload.EndpointID,
		}

		stack = edgeStack
	})
	if handler.DataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find a stack with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to persist the stack changes inside the database", err)
	}

	return response.JSON(w, stack)
}
