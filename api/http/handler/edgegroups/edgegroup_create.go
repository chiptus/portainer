package edgegroups

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/endpointutils"

	"github.com/asaskevich/govalidator"
)

type edgeGroupCreatePayload struct {
	Name         string
	Dynamic      bool
	TagIDs       []portaineree.TagID
	Endpoints    []portaineree.EndpointID
	PartialMatch bool
}

func (payload *edgeGroupCreatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("invalid Edge group name")
	}
	if payload.Dynamic && (payload.TagIDs == nil || len(payload.TagIDs) == 0) {
		return errors.New("tagIDs is mandatory for a dynamic Edge group")
	}
	if !payload.Dynamic && (payload.Endpoints == nil || len(payload.Endpoints) == 0) {
		return errors.New("environment is mandatory for a static Edge group")
	}
	return nil
}

// @id EdgeGroupCreate
// @summary Create an EdgeGroup
// @description **Access policy**: administrator
// @tags edge_groups
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body edgeGroupCreatePayload true "EdgeGroup data"
// @success 200 {object} portaineree.EdgeGroup
// @failure 503 "Edge compute features are disabled"
// @failure 500
// @router /edge_groups [post]
func (handler *Handler) edgeGroupCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload edgeGroupCreatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Edge groups from the database", err)
	}

	for _, edgeGroup := range edgeGroups {
		if edgeGroup.Name == payload.Name {
			return httperror.BadRequest("Edge group name must be unique", errors.New("edge group name must be unique"))
		}
	}

	edgeGroup := &portaineree.EdgeGroup{
		Name:         payload.Name,
		Dynamic:      payload.Dynamic,
		TagIDs:       []portaineree.TagID{},
		Endpoints:    []portaineree.EndpointID{},
		PartialMatch: payload.PartialMatch,
	}

	if edgeGroup.Dynamic {
		edgeGroup.TagIDs = payload.TagIDs
	} else {
		endpointIDs := []portaineree.EndpointID{}
		for _, endpointID := range payload.Endpoints {
			endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
			if err != nil {
				return httperror.InternalServerError("Unable to retrieve environment from the database", err)
			}

			if endpointutils.IsEdgeEndpoint(endpoint) {
				endpointIDs = append(endpointIDs, endpoint.ID)
			}
		}
		edgeGroup.Endpoints = endpointIDs
	}

	err = handler.DataStore.EdgeGroup().Create(edgeGroup)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the Edge group inside the database", err)
	}

	return response.JSON(w, edgeGroup)
}
