package endpointedge

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/portainer-ee/api/internal/edge"
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
// @router /endpoints/{id}/edge/generate-key [get]
func (handler *Handler) endpointEdgeGenerateKey(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var err error

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve settings from the database", err)
	}

	url := settings.EdgePortainerURL

	if url == "" {
		return httperror.BadRequest("Edge Portainer URL is not set", nil)
	}

	portainerHost, err := edge.ParseHostForEdge(url)
	if err != nil {
		return httperror.BadRequest("Unable to parse host", err)
	}

	edgeKey := handler.ReverseTunnelService.GenerateEdgeKey(url, portainerHost, 0)
	if err != nil {
		return httperror.InternalServerError("Unable to generate edge key", err)
	}

	return response.JSON(w, generateKeyResponse{EdgeKey: edgeKey})
}
