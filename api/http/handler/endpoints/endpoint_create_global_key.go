package endpoints

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
)

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

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the settings from the database", err)
	}

	// Create a new endpoint if none was found

	p := &endpointCreatePayload{
		Name:                 edgeID,
		URL:                  "https://" + r.Host,
		EndpointCreationType: edgeAgentEnvironment,
		GroupID:              1,
		TagIDs:               []portaineree.TagID{},
		EdgeCheckinInterval:  settings.EdgeAgentCheckinInterval,
		IsEdgeDevice:         true,
	}

	endpoint, hErr := handler.createEndpoint(p)
	if hErr != nil {
		return hErr
	}

	endpoint.UserTrusted = settings.TrustOnFirstConnect
	endpoint.EdgeID = edgeID

	err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return httperror.InternalServerError("Unable to persist environment changes inside the database", err)
	}

	relationObject := &portaineree.EndpointRelation{
		EndpointID: portaineree.EndpointID(endpoint.ID),
		EdgeStacks: map[portaineree.EdgeStackID]bool{},
	}

	err = handler.DataStore.EndpointRelation().Create(relationObject)
	if err != nil {
		return httperror.InternalServerError("Unable to persist the relation object inside the database", err)
	}

	handler.AuthorizationService.TriggerUsersAuthUpdate()

	return response.JSON(w, endpointCreateGlobalKeyResponse{endpoint.ID})
}
