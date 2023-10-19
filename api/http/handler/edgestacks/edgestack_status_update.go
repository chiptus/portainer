package edgestacks

import (
	"errors"
	"net/http"
	"slices"
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/asaskevich/govalidator"
	"github.com/rs/zerolog/log"
)

type updateStatusPayload struct {
	Error      string
	Status     *portainer.EdgeStackStatusType
	EndpointID portainer.EndpointID
	// RollbackTo specifies the stack file version to rollback to (only support to rollback to the last version currently)
	RollbackTo *int
	Time       int64
}

func (payload *updateStatusPayload) Validate(r *http.Request) error {
	if payload.Status == nil {
		return errors.New("invalid status")
	}

	if payload.EndpointID == 0 {
		return errors.New("invalid EnvironmentID")
	}

	if *payload.Status == portainer.EdgeStackStatusError && govalidator.IsNull(payload.Error) {
		return errors.New("error message is mandatory when status is error")
	}

	if payload.Time == 0 {
		payload.Time = time.Now().Unix()
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
// @param body body updateStatusPayload true "EdgeStack status payload"
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
	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		stack, err = handler.updateEdgeStackStatus(tx, r, portainer.EdgeStackID(stackID), payload)
		return err
	})
	if err != nil {
		var httpErr *httperror.HandlerError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		return httperror.InternalServerError("Unexpected error", err)
	}

	return response.JSON(w, stack)
}

func (handler *Handler) updateEdgeStackStatus(tx dataservices.DataStoreTx, r *http.Request, stackID portainer.EdgeStackID, payload updateStatusPayload) (*portaineree.EdgeStack, error) {
	stack, err := tx.EdgeStack().EdgeStack(stackID)
	if err != nil {
		if dataservices.IsErrObjectNotFound(err) {
			// skip error because agent tries to report on deleted stack
			log.Warn().
				Err(err).
				Int("stackID", int(stackID)).
				Int("status", int(*payload.Status)).
				Msg("Unable to find a stack inside the database, skipping error")
			return nil, nil
		}

		return nil, err
	}

	environmentStatus, ok := stack.Status[payload.EndpointID]
	if !ok {
		environmentStatus = portainer.EdgeStackStatus{
			EndpointID: portainer.EndpointID(payload.EndpointID),
			Status:     []portainer.EdgeStackDeploymentStatus{},
		}
	}

	// if the stack represents a successful remote update - skip it
	if slices.ContainsFunc(environmentStatus.Status, func(sts portainer.EdgeStackDeploymentStatus) bool {
		return sts.Type == portainer.EdgeStackStatusRemoteUpdateSuccess
	}) {
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

	log.Debug().
		Int("stackID", int(stackID)).
		Int("endpointID", int(payload.EndpointID)).
		Int("status", int(status)).
		Msg("Updating stack status")

	if stack.EdgeUpdateID != 0 {
		if status == portainer.EdgeStackStatusError {
			err := handler.edgeUpdateService.RemoveActiveSchedule(payload.EndpointID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
			if err != nil {
				log.Warn().
					Err(err).
					Msg("Failed to remove active schedule")
			}
		}

		if status == portainer.EdgeStackStatusRunning {
			handler.edgeUpdateService.EdgeStackDeployed(endpoint.ID, edgetypes.UpdateScheduleID(stack.EdgeUpdateID))
		}
	}

	updateEnvStatus(stack, environmentStatus, status, payload)

	err = tx.EdgeStack().UpdateEdgeStack(stackID, stack, true)
	if err != nil {
		return nil, handler.handlerDBErr(err, "Unable to persist the stack changes inside the database")
	}

	// stagger configuration check
	if handler.staggerService != nil &&
		stack.StaggerConfig != nil &&
		stack.StaggerConfig.StaggerOption != portaineree.EdgeStaggerOptionAllAtOnce {
		// StackFileVersion is used to differentiate the stagger workflow for the same edge stack
		handler.staggerService.UpdateStaggerEndpointStatusIfNeeds(stackID, stack.StackFileVersion, payload.RollbackTo, payload.EndpointID, status)
	}

	return stack, nil
}

func updateEnvStatus(edgeStack *portaineree.EdgeStack, environmentStatus portainer.EdgeStackStatus, status portainer.EdgeStackStatusType, payload updateStatusPayload) {
	if status == portainer.EdgeStackStatusRemoved {
		delete(edgeStack.Status, payload.EndpointID)
		return
	}

	if status == portainer.EdgeStackStatusAcknowledged {
		environmentStatus.Status = nil
	}

	environmentStatus.Status = append(environmentStatus.Status, portainer.EdgeStackDeploymentStatus{
		Type:  status,
		Error: payload.Error,
		Time:  payload.Time,
	})

	if status == portainer.EdgeStackStatusRunning {
		if payload.RollbackTo != nil && edgeStack.PreviousDeploymentInfo != nil {
			if edgeStack.PreviousDeploymentInfo.FileVersion == *payload.RollbackTo {
				log.Debug().Int("rollbackTo", *payload.RollbackTo).
					Int("endpointID", int(payload.EndpointID)).
					Msg("[stagger status update] rollback to the previous version")
				// if the endpoint is rolled back successfully, we should update the endpoint's edge
				// status's deploymentInfo to the previous version.
				environmentStatus.DeploymentInfo = portainer.StackDeploymentInfo{
					// !important. We should set the version as same as file version for rollback
					Version:     edgeStack.PreviousDeploymentInfo.FileVersion,
					FileVersion: edgeStack.PreviousDeploymentInfo.FileVersion,
					ConfigHash:  edgeStack.PreviousDeploymentInfo.ConfigHash,
				}

				environmentStatus.Status = insertRollbackEndpointStatus(environmentStatus.Status)
				edgeStack.Status[payload.EndpointID] = environmentStatus
				return
			}

			if edgeStack.StackFileVersion != *payload.RollbackTo {
				log.Debug().Int("rollbackTo", *payload.RollbackTo).
					Int("previousVersion", edgeStack.PreviousDeploymentInfo.FileVersion).
					Msg("unsupported rollbackTo version, fallback to the latest version")
			}
		}

		gitHash := ""
		if edgeStack.GitConfig != nil {
			gitHash = edgeStack.GitConfig.ConfigHash
		}
		environmentStatus.DeploymentInfo = portainer.StackDeploymentInfo{
			Version:     edgeStack.Version,
			FileVersion: edgeStack.StackFileVersion,
			ConfigHash:  gitHash,
		}
	}

	edgeStack.Status[payload.EndpointID] = environmentStatus
}

func insertRollbackEndpointStatus(environmentStatus []portainer.EdgeStackDeploymentStatus) []portainer.EdgeStackDeploymentStatus {
	if !slices.ContainsFunc(environmentStatus, func(sts portainer.EdgeStackDeploymentStatus) bool {
		return sts.Type == portainer.EdgeStackStatusRunning
	}) {
		// if the endpoint has no running status, we should not insert rollback status
		return environmentStatus
	}

	rolledBackTime := time.Now().Unix()
	if len(environmentStatus) > 0 {
		rolledBackTime = environmentStatus[len(environmentStatus)-1].Time
	}

	updatedEnvStatus := []portainer.EdgeStackDeploymentStatus{}
	updatedEnvStatus = append(updatedEnvStatus, environmentStatus[:len(environmentStatus)-1]...)

	updatedEnvStatus = append(updatedEnvStatus, portainer.EdgeStackDeploymentStatus{
		Type:  portainer.EdgeStackStatusRolledBack,
		Error: "",
		Time:  rolledBackTime,
	})

	updatedEnvStatus = append(updatedEnvStatus, environmentStatus[len(environmentStatus)-1:]...)

	return updatedEnvStatus
}
