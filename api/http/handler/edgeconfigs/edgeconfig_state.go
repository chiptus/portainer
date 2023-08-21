package edgeconfigs

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	"github.com/portainer/portainer-ee/api/internal/slices"
)

type edgeConfigStateTransition struct {
	From portaineree.EdgeConfigStateType
	To   []portaineree.EdgeConfigStateType
}

// @id EdgeConfigState
// @summary Update the state of an Edge configuration
// @description Update the state of an Edge configuration.
// @tags edge_configs
// @param id path int true "Edge configuration identifier"
// @param state path int true "Edge configuration state"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 404 "Edge configuration not found"
// @failure 500 "Server error"
// @router /edge_configurations/{id}/{state} [put]
func (h *Handler) edgeConfigState(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeID := r.Header.Get(portaineree.PortainerAgentEdgeIDHeader)

	endpointID, ok := h.dataStore.Endpoint().EndpointIDByEdgeID(edgeID)
	if !ok {
		return httperror.BadRequest("Invalid edge identifier provided", nil)
	}

	edgeConfigID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge configuration identifier route variable", err)
	}

	state, err := request.RetrieveNumericRouteVariableValue(r, "state")
	if err != nil {
		return httperror.BadRequest("Invalid edge configuration state route variable", err)
	}

	err = h.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		return TransitionToState(tx, portaineree.EdgeConfigID(edgeConfigID), portaineree.EndpointID(endpointID), portaineree.EdgeConfigStateType(state))
	})
	if err != nil {
		return httperror.InternalServerError("Could not update the edge config state", err)
	}

	cache.Del(endpointID)

	return nil
}

func TransitionToState(
	tx dataservices.DataStoreTx,
	edgeConfigID portaineree.EdgeConfigID,
	endpointID portaineree.EndpointID,
	nextEnvState portaineree.EdgeConfigStateType) error {

	edgeConfigState, err := tx.EdgeConfigState().Read(endpointID)
	if err != nil {
		return err
	}

	currentEnvState, ok := edgeConfigState.States[portaineree.EdgeConfigID(edgeConfigID)]
	if !ok {
		return errors.New("current state not found for edge config")
	}

	if !validTransition(currentEnvState, nextEnvState) {
		return errors.New("invalid transition")
	}

	edgeConfigState.States[portaineree.EdgeConfigID(edgeConfigID)] = nextEnvState

	err = tx.EdgeConfigState().Update(endpointID, edgeConfigState)
	if err != nil {
		return err
	}

	edgeConfig, err := tx.EdgeConfig().Read(portaineree.EdgeConfigID(edgeConfigID))
	if err != nil {
		return err
	}

	if nextEnvState == portaineree.EdgeConfigFailureState {
		edgeConfig.State = portaineree.EdgeConfigFailureState

		return tx.EdgeConfig().Update(edgeConfig.ID, edgeConfig)
	} else if nextEnvState != portaineree.EdgeConfigIdleState {
		return nil
	}

	switch edgeConfig.State {
	// Saving | Updating | Idle | Failure -> Idle
	case portaineree.EdgeConfigSavingState, portaineree.EdgeConfigUpdatingState, portaineree.EdgeConfigIdleState, portaineree.EdgeConfigFailureState:
		if currentEnvState == portaineree.EdgeConfigDeletingState {
			delete(edgeConfigState.States, edgeConfig.ID)

			if err = tx.EdgeConfigState().Update(endpointID, edgeConfigState); err != nil {
				return err
			}

			// Environment being removed indirectly
			if edgeConfig.State == portaineree.EdgeConfigIdleState {
				edgeConfig.Progress.Success--
			}

			edgeConfig.Progress.Total--
		} else if currentEnvState == portaineree.EdgeConfigSavingState || currentEnvState == portaineree.EdgeConfigUpdatingState {
			edgeConfig.Progress.Success++
		}

		if edgeConfig.Progress.Success == edgeConfig.Progress.Total {
			edgeConfig.State = portaineree.EdgeConfigIdleState
		}

	// Deleting -> Deleted
	case portaineree.EdgeConfigDeletingState:
		edgeConfig.Progress.Success++

		if edgeConfig.Progress.Success != edgeConfig.Progress.Total {
			break
		}

		if err = removeEdgeConfigStates(tx, edgeConfig.ID); err != nil {
			return err
		}

		return tx.EdgeConfig().Delete(edgeConfig.ID)
	}

	return tx.EdgeConfig().Update(edgeConfig.ID, edgeConfig)
}

func validTransition(current, next portaineree.EdgeConfigStateType) bool {
	idleOrFailure := []portaineree.EdgeConfigStateType{
		portaineree.EdgeConfigIdleState,
		portaineree.EdgeConfigFailureState,
	}

	transitions := []edgeConfigStateTransition{
		// Idle -> Saving | Updating | Deleting
		{
			From: portaineree.EdgeConfigIdleState,
			To: []portaineree.EdgeConfigStateType{
				portaineree.EdgeConfigSavingState,
				portaineree.EdgeConfigUpdatingState,
				portaineree.EdgeConfigDeletingState,
			},
		},

		// Saving -> Idle | Failure
		{
			From: portaineree.EdgeConfigSavingState,
			To:   idleOrFailure,
		},

		// Updating -> Idle | Failure
		{
			From: portaineree.EdgeConfigUpdatingState,
			To:   idleOrFailure,
		},

		// Deleting -> Idle | Failure
		{
			From: portaineree.EdgeConfigDeletingState,
			To:   idleOrFailure,
		},
	}

	for _, t := range transitions {
		if t.From == current && slices.Contains(t.To, next) {
			return true
		}
	}

	return false
}

func removeEdgeConfigStates(tx dataservices.DataStoreTx, edgeConfigID portaineree.EdgeConfigID) error {
	edgeConfigStates, err := tx.EdgeConfigState().ReadAll()
	if err != nil {
		return err
	}

	for _, edgeConfigState := range edgeConfigStates {
		if _, ok := edgeConfigState.States[edgeConfigID]; ok {
			delete(edgeConfigState.States, edgeConfigID)

			if err := tx.EdgeConfigState().Update(edgeConfigState.EndpointID, &edgeConfigState); err != nil {
				return err
			}
		}
	}

	return nil
}
