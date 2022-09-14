package licenses

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type (
	deletePayload struct {
		// List of license keys to remove
		LicenseKeys []string
	}

	deleteResponse struct {
		FailedKeys map[string]string `json:"failedKeys"`
	}
)

func (payload *deletePayload) Validate(r *http.Request) error {
	if payload.LicenseKeys == nil || len(payload.LicenseKeys) == 0 {
		return errors.New("Missing licenses keys")
	}

	return nil
}

// @id licensesDelete
// @summary delete license from portainer instance
// @description
// @description **Access policy**: administrator
// @tags license
// @security ApiKeyAuth
// @security jwt
// @accept json
// @produce json
// @param body body deletePayload true "list of license keys to remove"
// @success 200 {object} deleteResponse "Failures will be in `body.FailedKeys[key] = error`"
// @router /licenses [delete]
func (handler *Handler) licensesDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload deletePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	resp := &attachResponse{
		FailedKeys: map[string]string{},
	}

	for _, licenseKey := range payload.LicenseKeys {
		err := handler.LicenseService.DeleteLicense(licenseKey)
		if err != nil {
			resp.FailedKeys[licenseKey] = err.Error()
		}
	}

	return response.JSON(w, resp)
}
