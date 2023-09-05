package settings

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type settingsExperimentalInspectResponse struct {
	ExperimentalFeatures portaineree.ExperimentalFeatures `json:"experimentalFeatures"`
}

// @id SettingsExperimentalInspect
// @summary Retrieve Portainer experimental settings
// @description Retrieve Portainer experimental settings.
// @description **Access policy**: authenticated
// @tags settings
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {object} settingsExperimentalInspectResponse "Success"
// @failure 500 "Server error"
// @router /settings/experimental [get]
func (handler *Handler) settingsExperimentalInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the settings from the database", err)
	}

	expSettings := settingsExperimentalInspectResponse{
		ExperimentalFeatures: settings.ExperimentalFeatures,
	}

	return response.JSON(w, expSettings)
}
