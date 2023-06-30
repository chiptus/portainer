package edgestacks

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/pkg/featureflags"

	"github.com/asaskevich/govalidator"
	"github.com/rs/zerolog/log"
)

type updateStatusPayload struct {
	Error      string
	Status     *portainer.EdgeStackStatusType
	EndpointID portaineree.EndpointID
}

func (payload *updateStatusPayload) Validate(r *http.Request) error {
	if payload.Status == nil {
		return errors.New("Invalid status")
	}

	if payload.EndpointID == 0 {
		return errors.New("Invalid EnvironmentID")
	}

	if *payload.Status == portainer.EdgeStackStatusError && govalidator.IsNull(payload.Error) {
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
// @param id path int true "EdgeStack Id"
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

	var payload updateStatusPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var stack *portaineree.EdgeStack
	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		stack, err = handler.updateEdgeStackStatus(handler.DataStore, r, portaineree.EdgeStackID(stackID), payload)
	} else {
		err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
			stack, err = handler.updateEdgeStackStatus(tx, r, portaineree.EdgeStackID(stackID), payload)
			return err
		})
	}

	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, stack)
}

func (handler *Handler) updateEdgeStackStatus(tx dataservices.DataStoreTx, r *http.Request, stackID portaineree.EdgeStackID, payload updateStatusPayload) (*portaineree.EdgeStack, error) {
	stack, err := tx.EdgeStack().EdgeStack(stackID)
	if err != nil {
		return nil, handler.handlerDBErr(err, "Unable to find a stack with the specified identifier inside the database")
	}

	// if the stack represents a successful remote update - skip it
	if endpointStatus, ok := stack.Status[payload.EndpointID]; ok && endpointStatus.Details.RemoteUpdateSuccess {
		return stack, nil
	}

	endpoint, err := tx.Endpoint().Endpoint(payload.EndpointID)
	if err != nil {
		return nil, handler.handlerDBErr(err, "Unable to find an environment with the specified identifier inside the database")
	}

	err = handler.requestBouncer.AuthorizedEdgeEndpointOperation(r, endpoint)
	if err != nil {
		return nil, httperror.Forbidden("Permission denied to access environment", err)
	}

	status := *payload.Status

	if stack.EdgeUpdateID != 0 {
		if status == portainer.EdgeStackStatusError {
			err := handler.edgeUpdateService.RemoveActiveSchedule(payload.EndpointID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
			if err != nil {
				log.Warn().
					Err(err).
					Msg("Failed to remove active schedule")
			}
		}

		if status == portainer.EdgeStackStatusOk {
			handler.edgeUpdateService.EdgeStackDeployed(endpoint.ID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
		}
	}

	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		err = tx.EdgeStack().UpdateEdgeStackFunc(stackID, func(edgeStack *portaineree.EdgeStack) {
			details := edgeStack.Status[payload.EndpointID].Details
			details.Pending = false

			var deploymentInfo portainer.StackDeploymentInfo
			switch status {
			case portainer.EdgeStackStatusOk:
				details.Ok = true

				// if the edge stack is deployed successfully, we need to update the deployment info
				deploymentInfo.Version = edgeStack.Version
				if edgeStack.GitConfig != nil {
					deploymentInfo.ConfigHash = edgeStack.GitConfig.ConfigHash
				}
			case portainer.EdgeStackStatusError:
				details.Error = true
			case portainer.EdgeStackStatusAcknowledged:
				details.Acknowledged = true
			case portainer.EdgeStackStatusRemove:
				details.Remove = true
			case portainer.EdgeStackStatusImagesPulled:
				details.ImagesPulled = true
			}

			edgeStack.Status[payload.EndpointID] = portainer.EdgeStackStatus{
				Details:        details,
				Error:          payload.Error,
				EndpointID:     portainer.EndpointID(payload.EndpointID),
				DeploymentInfo: deploymentInfo,
			}

			stack = edgeStack
		})
		if err != nil {
			return nil, handler.handlerDBErr(err, "Unable to persist the stack changes inside the database")
		}
	} else {
		details := stack.Status[payload.EndpointID].Details
		details.Pending = false
		var deploymentInfo portainer.StackDeploymentInfo
		switch status {
		case portainer.EdgeStackStatusOk:
			details.Ok = true

			deploymentInfo.Version = stack.Version
			if stack.GitConfig != nil {
				deploymentInfo.ConfigHash = stack.GitConfig.ConfigHash
			}
		case portainer.EdgeStackStatusError:
			details.Error = true
		case portainer.EdgeStackStatusAcknowledged:
			details.Acknowledged = true
		case portainer.EdgeStackStatusRemove:
			details.Remove = true
		case portainer.EdgeStackStatusImagesPulled:
			details.ImagesPulled = true
		}

		stack.Status[payload.EndpointID] = portainer.EdgeStackStatus{
			Details:        details,
			Error:          payload.Error,
			EndpointID:     portainer.EndpointID(payload.EndpointID),
			DeploymentInfo: deploymentInfo,
		}

		err = tx.EdgeStack().UpdateEdgeStack(stackID, stack)
		if err != nil {
			return nil, handler.handlerDBErr(err, "Unable to persist the stack changes inside the database")
		}
	}

	return stack, nil
}
