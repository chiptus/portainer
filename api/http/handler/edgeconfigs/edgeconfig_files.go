package edgeconfigs

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer/api/filesystem"
)

type edgeConfigFilesPayload struct {
	ID         portaineree.EdgeConfigID `json:"id"`
	Name       string                   `json:"name"`
	BaseDir    string                   `json:"baseDir"`
	DirEntries []filesystem.DirEntry    `json:"dirEntries"`
	Prev       *edgeConfigFilesPayload  `json:"prev,omitempty"`
}

// @id EdgeConfigFiles
// @summary Get the files for an Edge configuration
// @description Retrieve the files for an Edge configuration.
// @description **Access policy**: authenticated
// @tags edge_configs
// @security ApiKeyAuth
// @security jwt
// @param id path int true "Edge configuration identifier"
// @success 200 {object} string "Success"
// @failure 400 "Invalid request"
// @failure 404 "Edge configuration not found"
// @failure 500 "Server error"
// @router /edge_configurations/{id}/files [get]
func (h *Handler) edgeConfigFiles(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeID := r.Header.Get(portaineree.PortainerAgentEdgeIDHeader)

	endpointID, ok := h.dataStore.Endpoint().EndpointIDByEdgeID(edgeID)
	if !ok {
		return httperror.BadRequest("Invalid edge identifier provided", nil)
	}

	edgeConfigID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return httperror.BadRequest("Invalid edge configuration identifier route variable", err)
	}

	edgeConfig, err := h.dataStore.EdgeConfig().Read(portaineree.EdgeConfigID(edgeConfigID))
	if h.dataStore.IsErrObjectNotFound(err) {
		return httperror.NotFound("Unable to find an edge configuration with the specified identifier inside the database", err)
	} else if err != nil {
		return httperror.InternalServerError("Unable to find an edge configuration with the specified identifier inside the database", err)
	}

	dirEntries, err := h.fileService.GetEdgeConfigDirEntries(edgeConfig, edgeID, portaineree.EdgeConfigCurrent)
	if err != nil {
		return httperror.InternalServerError("Unable to process the files for the edge configuration", err)
	}

	resp := edgeConfigFilesPayload{
		ID:         edgeConfig.ID,
		Name:       edgeConfig.Name,
		BaseDir:    edgeConfig.BaseDir,
		DirEntries: dirEntries,
	}

	edgeConfigState, err := h.dataStore.EdgeConfigState().Read(endpointID)
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the edge configuration state", err)
	}

	if edgeConfigState.States[edgeConfig.ID] == portaineree.EdgeConfigUpdatingState {
		prevDirEntries, err := h.fileService.GetEdgeConfigDirEntries(edgeConfig, edgeID, portaineree.EdgeConfigPrevious)
		if err != nil {
			return httperror.InternalServerError("Unable to process the files for the edge configuration", err)
		}

		resp.Prev = &edgeConfigFilesPayload{
			DirEntries: prevDirEntries,
		}
	}

	return response.JSON(w, resp)
}
