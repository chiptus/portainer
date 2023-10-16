package edgeconfigs

import (
	"errors"
	"net/http"
	"slices"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	"github.com/portainer/portainer-ee/api/internal/unique"
	portainer "github.com/portainer/portainer/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type edgeConfigUpdatePayload struct {
	Type         string
	EdgeGroupIDs []portainer.EdgeGroupID
}

func (p *edgeConfigUpdatePayload) Validate(r *http.Request) error {
	if _, ok := edgeConfigTypeMap[p.Type]; !ok {
		return errors.New("invalid type")
	}

	if len(p.EdgeGroupIDs) == 0 {
		return errors.New("edge group list cannot be empty")
	}

	return nil
}

// @id EdgeConfigUpdate
// @summary Update an Edge Configuration
// @description Update an Edge Configuration.
// @description **Access policy**: authenticated
// @tags edge_configs
// @security ApiKeyAuth
// @security jwt
// @accept multipart/form-data
// @produce json
// @param EdgeConfiguration formData edgeConfigUpdatePayload true "JSON stringified edgeConfigUpdatePayload object"
// @param File formData file true "File"
// @success 204
// @failure 400 "Invalid request"
// @router /edge_configurations/{id} [put]
func (h *Handler) edgeConfigUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeConfigID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var payload edgeConfigUpdatePayload
	err = request.RetrieveMultiPartFormJSONValue(r, "edgeConfiguration", &payload, false)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	if err := payload.Validate(r); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	file, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return httperror.BadRequest("Invalid request payload, missing file", err)
	}

	token, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.BadRequest("Invalid JWT token", err)
	}

	var endpointIDsToUpdate []portainer.EndpointID
	err = h.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		nextRelatedEndpointIDs, err := h.getRelatedEndpointIDs(tx, payload.EdgeGroupIDs)
		if err != nil {
			return err
		}

		edgeConfig, err := tx.EdgeConfig().Read(portaineree.EdgeConfigID(edgeConfigID))
		if err != nil {
			return err
		}

		currentRelatedEndpointIDs, err := h.getRelatedEndpointIDs(tx, edgeConfig.EdgeGroupIDs)
		if err != nil {
			return err
		}

		endpointIDsToUpdate = append(endpointIDsToUpdate, nextRelatedEndpointIDs...)
		endpointIDsToUpdate = append(endpointIDsToUpdate, currentRelatedEndpointIDs...)
		endpointIDsToUpdate = unique.Unique(endpointIDsToUpdate)

		if edgeConfig.State != portaineree.EdgeConfigIdleState && edgeConfig.State != portaineree.EdgeConfigFailureState {
			return errors.New("edge configuration cannot be updated while a deployment is in progress")
		}

		edgeConfig.Prev = &portaineree.EdgeConfig{
			Type:         edgeConfig.Type,
			Category:     edgeConfig.Category,
			EdgeGroupIDs: edgeConfig.EdgeGroupIDs,
		}

		edgeConfig.State = portaineree.EdgeConfigUpdatingState
		edgeConfig.Type = edgeConfigTypeMap[payload.Type]
		edgeConfig.EdgeGroupIDs = payload.EdgeGroupIDs
		edgeConfig.UpdatedBy = token.ID
		edgeConfig.Progress.Success = 0
		edgeConfig.Progress.Total = len(endpointIDsToUpdate)

		if err = tx.EdgeConfig().Update(edgeConfig.ID, edgeConfig); err != nil {
			return err
		}

		if len(file) > 0 {
			if err = h.fileService.RotateEdgeConfigs(edgeConfig.ID); err != nil {
				return err
			}

			if err = h.processEdgeConfigFile(edgeConfig.ID, file); err != nil {
				return err
			}
		}

		for _, endpointID := range endpointIDsToUpdate {
			endpoint, err := tx.Endpoint().Endpoint(endpointID)
			if err != nil {
				return httperror.BadRequest("Unable to retrieve endpoint", err)
			}

			// If it doesn't exist, create it
			edgeConfigState, err := tx.EdgeConfigState().Read(endpoint.ID)
			if err != nil {
				edgeConfigState = &portaineree.EdgeConfigState{
					EndpointID: endpoint.ID,
					States:     make(map[portaineree.EdgeConfigID]portaineree.EdgeConfigStateType),
				}

				if err := tx.EdgeConfigState().Create(edgeConfigState); err != nil {
					return httperror.InternalServerError("Unable to persist the edge configuration state inside the database", err)
				}
			}

			if s, ok := edgeConfigState.States[edgeConfig.ID]; ok &&
				(s != portaineree.EdgeConfigIdleState && s != portaineree.EdgeConfigFailureState) {
				return errors.New("edge configuration cannot be updated while a deployment is in progress")
			}

			if !slices.Contains(nextRelatedEndpointIDs, endpoint.ID) {
				edgeConfigState.States[edgeConfig.ID] = portaineree.EdgeConfigDeletingState
			} else if _, ok := edgeConfigState.States[edgeConfig.ID]; ok {
				edgeConfigState.States[edgeConfig.ID] = portaineree.EdgeConfigUpdatingState
			} else {
				edgeConfigState.States[edgeConfig.ID] = portaineree.EdgeConfigSavingState
			}

			if err = tx.EdgeConfigState().Update(endpoint.ID, edgeConfigState); err != nil {
				return httperror.InternalServerError("Unable to persist the edge configuration state inside the database", err)
			}

			if err = h.edgeAsyncService.PushConfigCommand(tx, endpoint, edgeConfig, edgeConfigState); err != nil {
				return httperror.InternalServerError("Unable to persist the edge configuration async command", err)
			}
		}

		return nil
	})
	if err != nil {
		return httperror.BadRequest("Unable to update the edge configuration in the database", err)
	}

	for _, endpointID := range endpointIDsToUpdate {
		cache.Del(endpointID)
	}

	return response.Empty(w)
}
