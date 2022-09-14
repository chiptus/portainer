package edgegroups

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/api/dataservices/errors"
)

// @id EdgeGroupInspect
// @summary Inspects an EdgeGroup
// @description **Access policy**: administrator
// @tags edge_groups
// @security ApiKeyAuth
// @security jwt
// @produce json
// @param id path int true "EdgeGroup Id"
// @success 200 {object} portaineree.EdgeGroup
// @failure 503 "Edge compute features are disabled"
// @failure 500
// @router /edge_groups/{id} [get]
func (handler *Handler) edgeGroupInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid Edge group identifier route variable", err)
	}

	edgeGroup, err := handler.DataStore.EdgeGroup().EdgeGroup(portaineree.EdgeGroupID(edgeGroupID))
	if err == errors.ErrObjectNotFound {
		return httperror.NotFound("Unable to find an Edge group with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an Edge group with the specified identifier inside the database", err)
	}

	if edgeGroup.Dynamic {
		endpoints, err := handler.getEndpointsByTags(edgeGroup.TagIDs, edgeGroup.PartialMatch)
		if err != nil {
			return httperror.InternalServerError("Unable to retrieve environments and environment groups for Edge group", err)
		}

		edgeGroup.Endpoints = endpoints
	}

	return response.JSON(w, edgeGroup)
}
