package licenses

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type (
	attachPayload struct {
		Key string `json:"key"`
	}

	attachResponse struct {
		ConflictingKeys []string `json:"conflictingKeys"`
	}
)

func (payload *attachPayload) Validate(r *http.Request) error {
	if payload == nil {
		return errors.New("missing licenses key")
	}

	return nil
}

// @id licensesAttach
// @summary attaches a list of licenses to Portainer
// @description
// @description **Access policy**: administrator
// @tags license
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body attachPayload true "list of licenses keys to attach"
// @param force query bool false "remove conflicting licenses"
// @success 200 {object} attachResponse "Success license data will be in `body.Licenses`, Failures will be in `body.ConflictingKeys = error`"
// @router /licenses/add [post]
func (handler *Handler) licensesAttach(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload attachPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	force, err := request.RetrieveBooleanQueryParameter(r, "force", true)
	if err != nil {
		return httperror.BadRequest("Failed parsing \"force\" boolean", err)
	}

	conflicts, err := handler.LicenseService.AddLicense(string(payload.Key), force)
	if err != nil {
		return httperror.BadRequest("License is invalid", err)
	}
	resp := attachResponse{
		ConflictingKeys: conflicts,
	}
	return response.JSON(w, resp)
}
