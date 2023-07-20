package endpointedge

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

type generateKeyResponse struct {
	EdgeKey string `json:"edgeKey"`
}

// endpointEdgeGenerateKey
// @summary Generate an EdgeKey
// @description Generates a general edge key
// @description **Access policy**: admin
// @tags edge, endpoints
// @accept json
// @produce json
// @param body body generateKeyResponse true "Generate Key Info"
// @success 200
// @failure 500
// @failure 400
// @router /endpoints/edge/generate-key [post]
func (handler *Handler) endpointEdgeGenerateKey(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var err error

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	apiURL := settings.EdgePortainerURL
	if apiURL == "" {
		return httperror.BadRequest("Portainer API server URL is not set in Edge Compute settings", nil)
	}

	tunnelAddr := settings.Edge.TunnelServerAddress

	edgeKey := handler.ReverseTunnelService.GenerateEdgeKey(apiURL, tunnelAddr, 0)
	if err != nil {
		return httperror.InternalServerError("Unable to generate edge key", err)
	}

	return response.JSON(w, generateKeyResponse{EdgeKey: edgeKey})
}
