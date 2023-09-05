package edgeconfigs

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id EdgeConfigDelete
// @summary Delete an Edge configuration
// @description Delete an Edge configuration.
// @description **Access policy**: authenticated
// @tags edge_configs
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Edge configuration identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 404 "Edge configuration not found"
// @failure 500 "Server error"
// @router /edge_configurations/{id} [delete]
func (h *Handler) edgeConfigDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeConfigID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge configuration identifier route variable", err)
	}

	var relatedEndpointIDs []portaineree.EndpointID

	err = h.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		relatedEndpointIDs, err = h.transitionToState(tx, portaineree.EdgeConfigID(edgeConfigID), portaineree.EdgeConfigDeletingState)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return httperror.BadRequest("Unable to delete edge configuration", err)
	}

	for _, endpointID := range relatedEndpointIDs {
		cache.Del(endpointID)
	}

	return response.Empty(w)
}

func (h *Handler) transitionToState(tx dataservices.DataStoreTx, edgeConfigID portaineree.EdgeConfigID, state portaineree.EdgeConfigStateType) ([]portaineree.EndpointID, error) {
	var relatedEndpointIDs []portaineree.EndpointID

	edgeConfig, err := tx.EdgeConfig().Read(portaineree.EdgeConfigID(edgeConfigID))
	if tx.IsErrObjectNotFound(err) {
		return nil, httperror.NotFound("Unable to find an edge configuration with the specified identifier inside the database", err)
	}

	if edgeConfig.State != portaineree.EdgeConfigIdleState {
		return nil, httperror.BadRequest("edge configuration cannot be updated unless it has been succesfully deployed first", err)
	}

	relatedEndpointIDs, err = h.getRelatedEndpointIDs(tx, edgeConfig.EdgeGroupIDs)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve the related endpoint IDs", err)
	}

	for _, endpointID := range relatedEndpointIDs {
		edgeConfigState, err := tx.EdgeConfigState().Read(endpointID)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve the edge configuration state", err)
		}

		edgeConfigState.States[edgeConfig.ID] = state

		tx.EdgeConfigState().Update(endpointID, edgeConfigState)

		endpoint, err := tx.Endpoint().Endpoint(portaineree.EndpointID(endpointID))
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve the endpoint", err)
		}

		if !endpoint.Edge.AsyncMode {
			continue
		}

		dirEntries, err := h.fileService.GetEdgeConfigDirEntries(edgeConfig, endpoint.EdgeID, portaineree.EdgeConfigCurrent)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to process the files for the edge configuration", err)
		}

		if err = h.edgeAsyncService.DeleteConfigCommandTx(tx, endpoint.ID, edgeConfig, dirEntries); err != nil {
			return nil, httperror.InternalServerError("Unable to add the edge config command", err)
		}
	}

	if len(relatedEndpointIDs) == 0 {
		return relatedEndpointIDs, tx.EdgeConfig().Delete(edgeConfig.ID)
	}

	edgeConfig.State = portaineree.EdgeConfigDeletingState
	edgeConfig.Progress.Success = 0

	err = tx.EdgeConfig().Update(edgeConfig.ID, edgeConfig)
	if err != nil {
		return nil, httperror.InternalServerError("Unable to find an edge configuration with the specified identifier inside the database", err)
	}

	return relatedEndpointIDs, nil
}
