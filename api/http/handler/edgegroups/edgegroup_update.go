package edgegroups

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"github.com/portainer/portainer-ee/api/internal/slices"
	"github.com/portainer/portainer-ee/api/internal/unique"

	"github.com/asaskevich/govalidator"
)

type edgeGroupUpdatePayload struct {
	Name         string
	Dynamic      bool
	TagIDs       []portaineree.TagID
	Endpoints    []portaineree.EndpointID
	PartialMatch *bool
}

func (payload *edgeGroupUpdatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("invalid Edge group name")
	}

	if payload.Dynamic && len(payload.TagIDs) == 0 {
		return errors.New("tagIDs is mandatory for a dynamic Edge group")
	}

	return nil
}

// @id EgeGroupUpdate
// @summary Updates an EdgeGroup
// @description **Access policy**: administrator
// @tags edge_groups
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param id path int true "EdgeGroup Id"
// @param body body edgeGroupUpdatePayload true "EdgeGroup data"
// @success 200 {object} portaineree.EdgeGroup
// @failure 503 "Edge compute features are disabled"
// @failure 500
// @router /edge_groups/{id} [put]
func (handler *Handler) edgeGroupUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Edge group identifier route variable", err)
	}

	var payload edgeGroupUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	var edgeGroup *portaineree.EdgeGroup
	err = handler.DataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		edgeGroup, err = tx.EdgeGroup().Read(portaineree.EdgeGroupID(edgeGroupID))
		if handler.DataStore.IsErrObjectNotFound(err) {
			return httperror.NotFound("Unable to find an Edge group with the specified identifier inside the database", err)
		} else if err != nil {
			return httperror.InternalServerError("Unable to find an Edge group with the specified identifier inside the database", err)
		}

		edgeGroups, err := tx.EdgeGroup().ReadAll()
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve Edge groups from the database", err)
		}

		if payload.Name != "" {
			for _, edgeGroup := range edgeGroups {
				if edgeGroup.Name == payload.Name && edgeGroup.ID != portaineree.EdgeGroupID(edgeGroupID) {
					return httperror.BadRequest("Edge group name must be unique", errors.New("edge group name must be unique"))
				}
			}

			edgeGroup.Name = payload.Name
		}

		endpoints, err := tx.Endpoint().Endpoints()
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve environments from database", err)
		}

		endpointGroups, err := tx.EndpointGroup().ReadAll()
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve environment groups from database", err)
		}

		oldRelatedEndpoints := edge.EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)

		edgeGroup.Dynamic = payload.Dynamic
		if err := calculateEndpointsOrTags(tx, edgeGroup, payload.Endpoints, payload.TagIDs); err != nil {
			return err
		}

		if payload.PartialMatch != nil {
			edgeGroup.PartialMatch = *payload.PartialMatch
		}

		err = tx.EdgeGroup().Update(edgeGroup.ID, edgeGroup)
		if err != nil {
			return httperror.InternalServerError("Unable to persist Edge group changes inside the database", err)
		}

		newRelatedEndpoints := edge.EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)
		endpointsToUpdate := unique.Unique(append(newRelatedEndpoints, oldRelatedEndpoints...))

		edgeJobs, err := tx.EdgeJob().ReadAll()
		if err != nil {
			return httperror.InternalServerError("Unable to fetch Edge jobs", err)
		}

		edgeStacks, err := tx.EdgeStack().EdgeStacks()
		if err != nil {
			return err
		}

		// Update the edgeGroups with the modified edgeGroup for updateEndpointStacks()
		for i := range edgeGroups {
			if edgeGroups[i].ID == edgeGroup.ID {
				edgeGroups[i] = *edgeGroup
			}
		}

		for _, endpointID := range endpointsToUpdate {
			endpoint, err := tx.Endpoint().Endpoint(endpointID)
			if err != nil {
				return httperror.InternalServerError("Unable to get Environment from database", err)
			}

			err = handler.updateEndpointStacks(tx, endpoint, edgeGroups, edgeStacks)
			if err != nil {
				return httperror.InternalServerError("Unable to persist Environment relation changes inside the database", err)
			}

			if !endpointutils.IsEdgeEndpoint(endpoint) {
				continue
			}

			var operation string
			if slices.Contains(newRelatedEndpoints, endpointID) && slices.Contains(oldRelatedEndpoints, endpointID) {
				continue
			} else if slices.Contains(newRelatedEndpoints, endpointID) {
				operation = "add"
			} else if slices.Contains(oldRelatedEndpoints, endpointID) {
				operation = "remove"
			} else {
				continue
			}

			if err = handler.updateEdgeConfigs(tx, endpoint, edgeGroup, operation); err != nil {
				return err
			}

			err = handler.updateEndpointEdgeJobs(tx, edgeGroup.ID, endpoint, edgeJobs, operation)
			if err != nil {
				return httperror.InternalServerError("Unable to persist Environment Edge Jobs changes inside the database", err)
			}
		}

		return nil
	})

	return txResponse(w, edgeGroup, err)
}

func (handler *Handler) updateEndpointStacks(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeGroups []portaineree.EdgeGroup, edgeStacks []portaineree.EdgeStack) error {
	relation, err := tx.EndpointRelation().EndpointRelation(endpoint.ID)
	if err != nil {
		return err
	}

	endpointGroup, err := tx.EndpointGroup().Read(endpoint.GroupID)
	if err != nil {
		return err
	}

	edgeStackSet := map[portaineree.EdgeStackID]bool{}

	endpointEdgeStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
	for _, edgeStackID := range endpointEdgeStacks {
		edgeStackSet[edgeStackID] = true
	}

	needsUpdate := false

	// If the edge stack is not related anymore, remove it
	for edgeStackID := range relation.EdgeStacks {
		if _, ok := edgeStackSet[edgeStackID]; !ok {
			needsUpdate = true

			err := handler.edgeAsyncService.RemoveStackCommandTx(tx, endpoint.ID, edgeStackID)
			if err != nil {
				return err
			}
		}
	}

	// If the edge stack was not related before, add it
	for edgeStackID := range edgeStackSet {
		if _, ok := relation.EdgeStacks[edgeStackID]; !ok {
			needsUpdate = true

			err := handler.edgeAsyncService.AddStackCommandTx(tx, endpoint, edgeStackID, "")
			if err != nil {
				return err
			}
		}
	}

	if !needsUpdate {
		return nil
	}

	relation.EdgeStacks = edgeStackSet

	return tx.EndpointRelation().UpdateEndpointRelation(endpoint.ID, relation)
}

func (handler *Handler) updateEndpointEdgeJobs(tx dataservices.DataStoreTx, edgeGroupID portaineree.EdgeGroupID, endpoint *portaineree.Endpoint, edgeJobs []portaineree.EdgeJob, operation string) error {
	for _, edgeJob := range edgeJobs {
		if !slices.Contains(edgeJob.EdgeGroups, edgeGroupID) {
			continue
		}

		switch operation {
		case "add":
			handler.ReverseTunnelService.AddEdgeJob(endpoint, &edgeJob)

			err := handler.edgeAsyncService.AddJobCommandTx(tx, endpoint.ID, edgeJob, []byte(edgeJob.ScriptPath))
			if err != nil {
				return err
			}
		case "remove":
			handler.ReverseTunnelService.RemoveEdgeJobFromEndpoint(endpoint.ID, edgeJob.ID)

			err := handler.edgeAsyncService.RemoveJobCommandTx(tx, endpoint.ID, edgeJob.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (handler *Handler) updateEdgeConfigs(tx dataservices.DataStoreTx, endpoint *portaineree.Endpoint, edgeGroup *portaineree.EdgeGroup, op string) error {
	edgeConfigs, err := tx.EdgeConfig().ReadAll()
	if err != nil {
		return err
	}

	var edgeConfigsToCreate []portaineree.EdgeConfigID
	for _, edgeConfig := range edgeConfigs {
		if slices.Contains(edgeConfig.EdgeGroupIDs, edgeGroup.ID) {
			edgeConfigsToCreate = append(edgeConfigsToCreate, edgeConfig.ID)
		}
	}

	edgeConfigsToCreate = unique.Unique(edgeConfigsToCreate)

	for _, edgeConfigID := range edgeConfigsToCreate {
		edgeConfig, err := tx.EdgeConfig().Read(edgeConfigID)
		if err != nil {
			return err
		}

		switch edgeConfig.State {
		case portaineree.EdgeConfigFailureState, portaineree.EdgeConfigDeletingState:
			continue
		}

		edgeConfigState, err := tx.EdgeConfigState().Read(endpoint.ID)
		if err != nil {
			edgeConfigState = &portaineree.EdgeConfigState{
				EndpointID: endpoint.ID,
				States:     make(map[portaineree.EdgeConfigID]portaineree.EdgeConfigStateType),
			}

			if err := tx.EdgeConfigState().Create(edgeConfigState); err != nil {
				return err
			}
		}

		switch op {
		case "add":
			edgeConfig.Progress.Total++

			if err := tx.EdgeConfig().Update(edgeConfigID, edgeConfig); err != nil {
				return err
			}

			edgeConfigState.States[edgeConfigID] = portaineree.EdgeConfigSavingState

		case "remove":
			edgeConfigState.States[edgeConfigID] = portaineree.EdgeConfigDeletingState
		}

		if err := tx.EdgeConfigState().Update(edgeConfigState.EndpointID, edgeConfigState); err != nil {
			return err
		}

		if !endpoint.Edge.AsyncMode {
			continue
		}

		dirEntries, err := handler.FileService.GetEdgeConfigDirEntries(edgeConfig, endpoint.EdgeID, portaineree.EdgeConfigCurrent)
		if err != nil {
			return httperror.InternalServerError("Unable to process the files for the edge configuration", err)
		}

		if err = handler.edgeAsyncService.AddConfigCommandTx(tx, endpoint.ID, edgeConfig, dirEntries); err != nil {
			return httperror.InternalServerError("Unable to persist the edge configuration command inside the database", err)
		}
	}

	cache.Del(endpoint.ID)

	return nil
}
