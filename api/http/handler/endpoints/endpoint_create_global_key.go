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

	// Search for existing endpoints for the given edgeID

	endpoints, err := handler.dataStore.Endpoint().Endpoints()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve the endpoints from the database", err}
	}

	for _, endpoint := range endpoints {
		if endpoint.EdgeID == edgeID {
			return response.JSON(w, endpointCreateGlobalKeyResponse{endpoint.ID})
		}
	}

	// Create a new endpoint if none was found

	p := &endpointCreatePayload{
		Name:                 edgeID,
		URL:                  "https://" + r.Host,
		EndpointCreationType: edgeAgentEnvironment,
		GroupID:              1,
		EdgeCheckinInterval:  portaineree.DefaultEdgeAgentCheckinIntervalInSeconds,
		IsEdgeDevice:         true,
	}

	endpoint, hErr := handler.createEndpoint(p)
	if hErr != nil {
		return hErr
	}

	settings, err := handler.dataStore.Settings().Settings()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve the settings from the database", err}
	}

	endpoint.UserTrusted = settings.TrustOnFirstConnect
	endpoint.EdgeID = edgeID

	err = handler.dataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist environment changes inside the database", err}
	}

	relationObject := &portaineree.EndpointRelation{
		EndpointID: portaineree.EndpointID(endpoint.ID),
		EdgeStacks: map[portaineree.EdgeStackID]bool{},
	}

	err = handler.dataStore.EndpointRelation().Create(relationObject)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the relation object inside the database", err}
	}

	handler.AuthorizationService.TriggerUsersAuthUpdate()

	return response.JSON(w, endpointCreateGlobalKeyResponse{endpoint.ID})
}
