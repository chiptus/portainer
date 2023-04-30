package settings

import (
	"net/http"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type settingsExperimentalUpdatePayload struct {
	// OpenAI integration
	OpenAIIntegration *bool `example:"true" validate:"required"`
}

func (payload *settingsExperimentalUpdatePayload) Validate(r *http.Request) error {
	if payload.OpenAIIntegration == nil {
		return errors.New("invalid OpenAIIntegration value")
	}
	return nil
}

// @id SettingsExperimentalUpdate
// @summary Update Portainer experimental settings
// @description Update Portainer experimental settings.
// @description **Access policy**: administrator
// @tags settings
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body settingsExperimentalUpdatePayload true "New settings"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /settings/experimental [put]
func (handler *Handler) settingsExperimentalUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload settingsExperimentalUpdatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the settings from the database", err)
	}

	settings.ExperimentalFeatures.OpenAIIntegration = *payload.OpenAIIntegration

	err = handler.DataStore.Settings().UpdateSettings(settings)
	if err != nil {
		return httperror.InternalServerError("Unable to persist settings changes inside the database", err)
	}

	return response.Empty(w)
}
