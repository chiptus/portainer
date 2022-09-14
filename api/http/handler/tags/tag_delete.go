package tags

import (
	"github.com/portainer/portainer-ee/api/internal/endpointutils"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer/api/dataservices/errors"
)

// @id TagDelete
// @summary Remove a tag
// @description Remove a tag.
// @description **Access policy**: administrator
// @tags tags
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Tag identifier"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Tag not found"
// @failure 500 "Server error"
// @router /tags/{id} [delete]
func (handler *Handler) tagDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	id, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid tag identifier route variable", err)
	}
	tagID := portaineree.TagID(id)

	tag, err := handler.DataStore.Tag().Tag(tagID)
	if err == errors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find a tag with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find a tag with the specified identifier inside the database", err)
	}

	for endpointID := range tag.Endpoints {
		endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve environment from the database", err)
		}

		tagIdx := findTagIndex(endpoint.TagIDs, tagID)
		if tagIdx != -1 {
			endpoint.TagIDs = removeElement(endpoint.TagIDs, tagIdx)
			err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
			if err != nil {
				return httperror.InternalServerError("Unable to update environment", err)
			}
		}
	}

	for endpointGroupID := range tag.EndpointGroups {
		endpointGroup, err := handler.DataStore.EndpointGroup().EndpointGroup(endpointGroupID)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve environment group from the database", err)
		}

		tagIdx := findTagIndex(endpointGroup.TagIDs, tagID)
		if tagIdx != -1 {
			endpointGroup.TagIDs = removeElement(endpointGroup.TagIDs, tagIdx)
			err = handler.DataStore.EndpointGroup().UpdateEndpointGroup(endpointGroup.ID, endpointGroup)
			if err != nil {
				return httperror.InternalServerError("Unable to update environment group", err)
			}
		}
	}

	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve environments from the database", err)
	}

	edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve edge groups from the database", err)
	}

	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve edge stacks from the database", err)
	}

	for _, endpoint := range endpoints {
		if (tag.Endpoints[endpoint.ID] || tag.EndpointGroups[endpoint.GroupID]) && (endpointutils.IsEdgeEndpoint(&endpoint)) {
			err = handler.updateEndpointRelations(endpoint, edgeGroups, edgeStacks)
			if err != nil {
				return httperror.InternalServerError("Unable to update environment relations in the database", err)
			}
		}
	}

	for idx := range edgeGroups {
		edgeGroup := &edgeGroups[idx]
		tagIdx := findTagIndex(edgeGroup.TagIDs, tagID)
		if tagIdx != -1 {
			edgeGroup.TagIDs = removeElement(edgeGroup.TagIDs, tagIdx)
			err = handler.DataStore.EdgeGroup().UpdateEdgeGroup(edgeGroup.ID, edgeGroup)
			if err != nil {
				return httperror.InternalServerError("Unable to update environment group", err)
			}
		}
	}

	err = handler.DataStore.Tag().DeleteTag(tagID)
	if err != nil {
		return httperror.InternalServerError("Unable to remove the tag from the database", err)
	}

	return response.Empty(w)
}

func (handler *Handler) updateEndpointRelations(endpoint portaineree.Endpoint, edgeGroups []portaineree.EdgeGroup, edgeStacks []portaineree.EdgeStack) error {
	endpointRelation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpoint.ID)
	if err != nil {
		return err
	}

	endpointGroup, err := handler.DataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
	if err != nil {
		return err
	}

	endpointStacks := edge.EndpointRelatedEdgeStacks(&endpoint, endpointGroup, edgeGroups, edgeStacks)
	stacksSet := map[portaineree.EdgeStackID]bool{}
	for _, edgeStackID := range endpointStacks {
		stacksSet[edgeStackID] = true
	}
	endpointRelation.EdgeStacks = stacksSet

	return handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpoint.ID, endpointRelation)
}

func findTagIndex(tags []portaineree.TagID, searchTagID portaineree.TagID) int {
	for idx, tagID := range tags {
		if searchTagID == tagID {
			return idx
		}
	}
	return -1
}

func removeElement(arr []portaineree.TagID, index int) []portaineree.TagID {
	if index < 0 {
		return arr
	}
	lastTagIdx := len(arr) - 1
	arr[index] = arr[lastTagIdx]
	return arr[:lastTagIdx]
}
