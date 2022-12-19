package edgegroups

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/internal/slices"
)

type decoratedEdgeGroup struct {
	portaineree.EdgeGroup
	HasEdgeStack  bool `json:"HasEdgeStack"`
	HasEdgeGroup  bool `json:"HasEdgeGroup"`
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
	edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Edge groups from the database", err)
	}

	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Edge stacks from the database", err)
	}

	usedEdgeGroups := make(map[portaineree.EdgeGroupID]bool)

	for _, stack := range edgeStacks {
		for _, groupID := range stack.EdgeGroups {
			usedEdgeGroups[groupID] = true
		}
	}

	edgeJobs, err := handler.DataStore.EdgeJob().EdgeJobs()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Edge jobs from the database", err)
	}

	decoratedEdgeGroups := []decoratedEdgeGroup{}
	for _, orgEdgeGroup := range edgeGroups {
		usedByEdgeJob := false
		for _, edgeJob := range edgeJobs {
			if slices.Contains(edgeJob.EdgeGroups, portaineree.EdgeGroupID(orgEdgeGroup.ID)) {
				usedByEdgeJob = true
				break
			}
		}

		edgeGroup := decoratedEdgeGroup{
			EdgeGroup:     orgEdgeGroup,
			EndpointTypes: []portaineree.EndpointType{},
		}
		if edgeGroup.Dynamic {
			endpointIDs, err := handler.getEndpointsByTags(edgeGroup.TagIDs, edgeGroup.PartialMatch)
			if err != nil {
				return httperror.InternalServerError("Unable to retrieve environments and environment groups for Edge group", err)
			}

			edgeGroup.Endpoints = endpointIDs
		}

		endpointTypes, err := getEndpointTypes(handler.DataStore.Endpoint(), edgeGroup.Endpoints)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve environment types for Edge group", err)
		}

		edgeGroup.EndpointTypes = endpointTypes

		edgeGroup.HasEdgeStack = usedEdgeGroups[edgeGroup.ID]

		edgeGroup.HasEdgeGroup = usedByEdgeJob

		decoratedEdgeGroups = append(decoratedEdgeGroups, edgeGroup)
	}

	return response.JSON(w, decoratedEdgeGroups)
}

func getEndpointTypes(endpointService dataservices.EndpointService, endpointIds []portaineree.EndpointID) ([]portaineree.EndpointType, error) {
	typeSet := map[portaineree.EndpointType]bool{}
	for _, endpointID := range endpointIds {
		endpoint, err := endpointService.Endpoint(endpointID)
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
