package edgegroups

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/slices"
	"github.com/portainer/portainer/pkg/featureflags"
)

type decoratedEdgeGroup struct {
	portaineree.EdgeGroup
	HasEdgeStack  bool `json:"HasEdgeStack"`
	HasEdgeJob    bool `json:"HasEdgeJob"`
	HasEdgeConfig bool `json:"HasEdgeConfig"`
	EndpointTypes []portaineree.EndpointType
}

// @id EdgeGroupList
// @summary list EdgeGroups
// @description **Access policy**: administrator
// @tags edge_groups
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} decoratedEdgeGroup "EdgeGroups"
// @failure 500
// @failure 503 "Edge compute features are disabled"
// @router /edge_groups [get]
func (handler *Handler) edgeGroupList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var decoratedEdgeGroups []decoratedEdgeGroup
	var err error

	if featureflags.IsEnabled(portaineree.FeatureNoTx) {
		decoratedEdgeGroups, err = getEdgeGroupList(handler.DataStore)
	} else {
		err = handler.DataStore.ViewTx(func(tx dataservices.DataStoreTx) error {
			decoratedEdgeGroups, err = getEdgeGroupList(tx)
			return err
		})
	}

	return txResponse(w, decoratedEdgeGroups, err)
}

func getEdgeGroupList(tx dataservices.DataStoreTx) ([]decoratedEdgeGroup, error) {
	edgeGroups, err := tx.EdgeGroup().ReadAll()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve Edge groups from the database", err)
	}

	edgeStacks, err := tx.EdgeStack().EdgeStacks()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve Edge stacks from the database", err)
	}

	usedEdgeGroups := make(map[portaineree.EdgeGroupID]bool)

	for _, stack := range edgeStacks {
		for _, groupID := range stack.EdgeGroups {
			usedEdgeGroups[groupID] = true
		}
	}

	edgeJobs, err := tx.EdgeJob().ReadAll()
	if err != nil {
		return nil, httperror.InternalServerError("Unable to retrieve Edge jobs from the database", err)
	}

	decoratedEdgeGroups := []decoratedEdgeGroup{}
	for _, orgEdgeGroup := range edgeGroups {
		// hide groups that are created for an edge update
		if orgEdgeGroup.EdgeUpdateID != 0 {
			continue
		}

		usedByEdgeJob := false
		for _, edgeJob := range edgeJobs {
			if slices.Contains(edgeJob.EdgeGroups, portaineree.EdgeGroupID(orgEdgeGroup.ID)) {
				usedByEdgeJob = true
				break
			}
		}

		usedByEdgeConfig := false
		edgeConfigs, err := tx.EdgeConfig().ReadAll()
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve Edge configs from the database", err)
		}

		for _, edgeConfig := range edgeConfigs {
			if slices.Contains(edgeConfig.EdgeGroupIDs, portaineree.EdgeGroupID(orgEdgeGroup.ID)) {
				usedByEdgeConfig = true
				break
			}
		}

		edgeGroup := decoratedEdgeGroup{
			EdgeGroup:     orgEdgeGroup,
			EndpointTypes: []portaineree.EndpointType{},
		}
		if edgeGroup.Dynamic {
			endpointIDs, err := GetEndpointsByTags(tx, edgeGroup.TagIDs, edgeGroup.PartialMatch)
			if err != nil {
				return nil, httperror.InternalServerError("Unable to retrieve environments and environment groups for Edge group", err)
			}

			edgeGroup.Endpoints = endpointIDs
		}

		endpointTypes, err := getEndpointTypes(tx, edgeGroup.Endpoints)
		if err != nil {
			return nil, httperror.InternalServerError("Unable to retrieve environment types for Edge group", err)
		}

		edgeGroup.EndpointTypes = endpointTypes
		edgeGroup.HasEdgeStack = usedEdgeGroups[edgeGroup.ID]
		edgeGroup.HasEdgeJob = usedByEdgeJob
		edgeGroup.HasEdgeConfig = usedByEdgeConfig

		decoratedEdgeGroups = append(decoratedEdgeGroups, edgeGroup)
	}

	return decoratedEdgeGroups, nil
}

func getEndpointTypes(tx dataservices.DataStoreTx, endpointIds []portaineree.EndpointID) ([]portaineree.EndpointType, error) {
	typeSet := map[portaineree.EndpointType]bool{}
	for _, endpointID := range endpointIds {
		endpoint, err := tx.Endpoint().Endpoint(endpointID)
		if err != nil {
			return nil, fmt.Errorf("failed fetching environment: %w", err)
		}

		typeSet[endpoint.Type] = true
	}

	endpointTypes := make([]portaineree.EndpointType, 0, len(typeSet))
	for endpointType := range typeSet {
		endpointTypes = append(endpointTypes, endpointType)
	}

	return endpointTypes, nil
}
