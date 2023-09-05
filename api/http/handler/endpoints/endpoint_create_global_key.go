package endpoints

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type endpointCreateGlobalKeyPayload struct {
	EdgeGroupsIDs      []portaineree.EdgeGroupID
	EnvironmentGroupID portaineree.EndpointGroupID
	TagsIDs            []portaineree.TagID
}

func (payload *endpointCreateGlobalKeyPayload) Validate(request *http.Request) error {
	if payload.EnvironmentGroupID == 0 {
		payload.EnvironmentGroupID = 1
	}

	return nil
}

type endpointCreateGlobalKeyResponse struct {
	EndpointID portaineree.EndpointID `json:"endpointID"`
}

// @id EndpointCreateGlobalKey
// @summary Create or retrieve the endpoint for an EdgeID
// @tags endpoints
// @success 200 {object} endpointCreateGlobalKeyResponse "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /endpoints/global-key [post]
func (handler *Handler) endpointCreateGlobalKey(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeID := r.Header.Get(portaineree.PortainerAgentEdgeIDHeader)
	if edgeID == "" {
		return httperror.BadRequest("Invalid Edge ID", errors.New("the Edge ID cannot be empty"))
	}

	err := handler.requestBouncer.AuthorizedClientTLSConn(r)
	if err != nil {
		return httperror.Forbidden("forbidden request", err)
	}

	// Search for existing endpoints for the given edgeID

	endpointID, ok := handler.DataStore.Endpoint().EndpointIDByEdgeID(edgeID)
	if ok {
		return response.JSON(w, endpointCreateGlobalKeyResponse{endpointID})
	}

	var payload *endpointCreateGlobalKeyPayload
	if r.Body != nil && r.ContentLength > 0 {
		payload, err = request.GetPayload[endpointCreateGlobalKeyPayload](r)
		if err != nil {
			return httperror.BadRequest("Invalid request payload", err)
		}
	} else {
		payload = &endpointCreateGlobalKeyPayload{
			EnvironmentGroupID: 1,
			EdgeGroupsIDs:      []portaineree.EdgeGroupID{},
			TagsIDs:            []portaineree.TagID{},
		}
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the settings from the database", err)
	}

	// validate the environment group
	_, err = handler.DataStore.EndpointGroup().Read(payload.EnvironmentGroupID)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to retrieve the environment group from the database")
		payload.EnvironmentGroupID = 1
	}

	// validate tags
	tagsIDs := []portaineree.TagID{}
	for _, tagID := range payload.TagsIDs {
		_, err := handler.DataStore.Tag().Read(tagID)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to retrieve the tag from the database")
			continue
		}

		tagsIDs = append(tagsIDs, tagID)
	}

	// validate edge groups
	var edgeGroupsIDs []portaineree.EdgeGroupID
	for _, edgeGroupID := range payload.EdgeGroupsIDs {
		_, err := handler.DataStore.EdgeGroup().Read(edgeGroupID)
		if err != nil {
			log.Warn().Err(err).Msg("Unable to retrieve the edge group from the database")
			continue
		}

		edgeGroupsIDs = append(edgeGroupsIDs, edgeGroupID)
	}

	// Create a new endpoint if none was found
	p := &endpointCreatePayload{
		Name:                 edgeID,
		EndpointCreationType: edgeAgentEnvironment,
		GroupID:              int(payload.EnvironmentGroupID),
		TagIDs:               tagsIDs,
		EdgeCheckinInterval:  settings.EdgeAgentCheckinInterval,
	}

	endpoint, hErr := handler.createEndpoint(handler.DataStore, p)
	if hErr != nil {
		return hErr
	}

	endpoint.UserTrusted = settings.TrustOnFirstConnect
	endpoint.EdgeID = edgeID

	err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to persist environment changes inside the database", err)
	}

	err = edge.AddEnvironmentToEdgeGroups(handler.DataStore, endpoint, edgeGroupsIDs)
	if err != nil {
		return httperror.InternalServerError("Unable to add environment to edge groups", err)
	}

	handler.AuthorizationService.TriggerUsersAuthUpdate()

	return response.JSON(w, endpointCreateGlobalKeyResponse{endpoint.ID})
}
