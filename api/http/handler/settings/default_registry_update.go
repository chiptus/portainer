package settings

import (
	"net/http"

	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type defaultRegistryUpdatePayload struct {
	Hide bool `json:"Hide,omitempty" example:"false"`
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
// @router /settings/default_registry [put]
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
