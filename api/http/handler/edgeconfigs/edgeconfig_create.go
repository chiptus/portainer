package edgeconfigs

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	"github.com/portainer/portainer-ee/api/dataservices/edgeconfig"
	"github.com/portainer/portainer-ee/api/http/security"
	"github.com/portainer/portainer-ee/api/internal/edge"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	"github.com/portainer/portainer-ee/api/internal/unique"
)

type edgeConfigCreatePayload struct {
	Name         string
	BaseDir      string
	Type         string
	EdgeGroupIDs []portaineree.EdgeGroupID
}

var edgeConfigTypeMap = map[string]portaineree.EdgeConfigType{
	"general":    edgeconfig.EdgeConfigTypeGeneral,
	"filename":   edgeconfig.EdgeConfigTypeSpecificFile,
	"foldername": edgeconfig.EdgeConfigTypeSpecificFolder,
}

func (p *edgeConfigCreatePayload) Validate(r *http.Request) error {
	if len(p.Name) == 0 {
		return errors.New("invalid name")
	}

	if len(p.BaseDir) == 0 || !filepath.IsAbs(p.BaseDir) {
		return errors.New("invalid directory path")
	}

	if _, ok := edgeConfigTypeMap[p.Type]; !ok {
		return errors.New("invalid type")
	}

	if len(p.EdgeGroupIDs) == 0 {
		return errors.New("edge group list cannot be empty")
	}

	return nil
}

// @id EdgeConfigCreate
// @summary Create an Edge Configuration
// @description Create an Edge Configuration.
// @description **Access policy**: authenticated
// @tags edge_configs
// @security ApiKeyAuth
// @security jwt
// @accept multipart/form-data
// @param EdgeConfiguration formData edgeConfigCreatePayload true "JSON stringified edgeConfigCreatePayload object"
// @param File formData file true "File"
// @success 204
// @failure 400 "Invalid request"
// @router /edge_configurations [post]
func (h *Handler) edgeConfigCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload edgeConfigCreatePayload
	err := request.RetrieveMultiPartFormJSONValue(r, "edgeConfiguration", &payload, false)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	if err := payload.Validate(r); err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	file, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return httperror.BadRequest("Invalid request payload, missing file", err)
	}

	token, err := security.RetrieveTokenData(r)
	if err != nil {
		return httperror.BadRequest("Invalid JWT token", err)
	}

	edgeConfig := &portaineree.EdgeConfig{
		Name:         payload.Name,
		BaseDir:      payload.BaseDir,
		Type:         edgeConfigTypeMap[payload.Type],
		State:        portaineree.EdgeConfigSavingState,
		EdgeGroupIDs: payload.EdgeGroupIDs,
		CreatedBy:    token.ID,
	}

	var relatedEndpointIDs []portaineree.EndpointID

	err = h.dataStore.UpdateTx(func(tx dataservices.DataStoreTx) error {
		relatedEndpointIDs, err = h.getRelatedEndpointIDs(tx, payload.EdgeGroupIDs)
		if err != nil {
			return httperror.BadRequest("Unable to retrieve related endpoints", err)
		}

		if len(relatedEndpointIDs) == 0 {
			edgeConfig.State = portaineree.EdgeConfigIdleState
		}

		edgeConfig.Progress.Total = len(relatedEndpointIDs)

		if err = tx.EdgeConfig().Create(edgeConfig); err != nil {
			return httperror.BadRequest("Unable to persist the edge configuration inside the database", err)
		}

		if err = h.processEdgeConfigFile(edgeConfig.ID, file); err != nil {
			return httperror.BadRequest("Unable to process the uploaded file", err)
		}

		for _, endpointID := range relatedEndpointIDs {
			endpoint, err := tx.Endpoint().Endpoint(endpointID)
			if err != nil {
				return httperror.BadRequest("Unable to retrieve endpoint", err)
			}

			// If it doesn't exist, create it
			edgeConfigState, err := tx.EdgeConfigState().Read(endpoint.ID)
			if err != nil {
				edgeConfigState = &portaineree.EdgeConfigState{
					EndpointID: endpoint.ID,
					States:     make(map[portaineree.EdgeConfigID]portaineree.EdgeConfigStateType),
				}

				if err := tx.EdgeConfigState().Create(edgeConfigState); err != nil {
					return httperror.InternalServerError("Unable to persist the edge configuration state inside the database", err)
				}
			}

			edgeConfigState.States[edgeConfig.ID] = portaineree.EdgeConfigSavingState

			if err = tx.EdgeConfigState().Update(endpoint.ID, edgeConfigState); err != nil {
				return httperror.InternalServerError("Unable to persist the edge configuration state inside the database", err)
			}

			if !endpoint.Edge.AsyncMode {
				continue
			}

			dirEntries, err := h.fileService.GetEdgeConfigDirEntries(edgeConfig, endpoint.EdgeID, portaineree.EdgeConfigCurrent)
			if err != nil {
				return httperror.InternalServerError("Unable to process the files for the edge configuration", err)
			}

			if err = h.edgeAsyncService.AddConfigCommandTx(tx, endpoint.ID, edgeConfig, dirEntries); err != nil {
				return httperror.InternalServerError("Unable to persist the edge configuration command inside the database", err)
			}
		}

		return nil
	})
	if err != nil {
		return httperror.BadRequest("Unable to persist the edge configuration inside the database", err)
	}

	for _, endpointID := range relatedEndpointIDs {
		cache.Del(endpointID)
	}

	return response.Empty(w)
}

func (h *Handler) getRelatedEndpointIDs(tx dataservices.DataStoreTx, edgeGroupIDs []portaineree.EdgeGroupID) ([]portaineree.EndpointID, error) {
	endpoints, err := tx.Endpoint().Endpoints()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve endpoints: %w", err)
	}

	n := 0
	for _, endpoint := range endpoints {
		if endpoint.Type == portaineree.EdgeAgentOnDockerEnvironment {
			endpoints[n] = endpoint
			n++
		}
	}
	endpoints = endpoints[:n]

	endpointGroups, err := tx.EndpointGroup().ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve endpoint groups: %w", err)
	}

	var relatedEndpointIDs []portaineree.EndpointID

	for _, edgeGroupID := range edgeGroupIDs {
		edgeGroup, err := tx.EdgeGroup().Read(edgeGroupID)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve edge group: %w", err)
		}

		relatedEndpointIDs = append(relatedEndpointIDs, edge.EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)...)
	}

	return unique.Unique(relatedEndpointIDs), nil
}

func (h *Handler) processEdgeConfigFile(edgeConfigID portaineree.EdgeConfigID, file []byte) error {
	zipFile, err := zip.NewReader(bytes.NewReader(file), int64(len(file)))
	if err != nil {
		return err
	}

	for _, f := range zipFile.File {
		if f.FileInfo().IsDir() {
			continue
		}

		rd, err := f.Open()
		if err != nil {
			return err
		}
		defer rd.Close()

		if err = h.fileService.StoreEdgeConfigFile(edgeConfigID, f.Name, rd); err != nil {
			return err
		}
	}

	return nil
}
