package settings

import (
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"net/http"
)

type defaultRegistryUpdatePayload struct {
	Hide bool
}

// @id DefaultRegistryUpdate
// @summary Update Portainer default registry settings
// @description Update Portainer default registry settings.
// @description **Access policy**: administrator
// @tags settings
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body defaultRegistryUpdatePayload true "Update default registry"
// @success 200 {object} portaineree.Settings "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /settings [put]
func (handler *Handler) defaultRegistryUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload defaultRegistryUpdatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve the settings from the database", err)
	}

	settings.DefaultRegistry = payload

	err = handler.DataStore.Settings().UpdateSettings(settings)
	if err != nil {
		return httperror.InternalServerError("Unable to persist settings changes inside the database", err)
	}

	return response.JSON(w, settings)
}

func (b defaultRegistryUpdatePayload) Validate(request *http.Request) error {
	return nil
}
