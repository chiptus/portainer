package licenses

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	"github.com/portainer/liblicense"
)

type (
	attachPayload struct {
		// List of license keys to attach
		LicenseKeys []string
	}

	attachResponse struct {
		Licenses   []*liblicense.PortainerLicense `json:"licenses"`
		FailedKeys map[string]string              `json:"failedKeys"`
	}
)

func (payload *attachPayload) Validate(r *http.Request) error {
	if payload.LicenseKeys == nil || len(payload.LicenseKeys) == 0 {
		return errors.New("Missing licenses keys")
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
// @success 200 {object} attachResponse "Success license data will be in `body.Licenses`, Failures will be in `body.FailedKeys[key] = error`"
// @router /licenses/attach [post]
func (handler *Handler) licensesAttach(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload attachPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	resp := &attachResponse{
		FailedKeys: map[string]string{},
		Licenses:   []*liblicense.PortainerLicense{},
	}

	for _, licenseKey := range payload.LicenseKeys {
		license, err := handler.LicenseService.AddLicense(licenseKey)
		if err != nil {
			resp.FailedKeys[licenseKey] = err.Error()
			continue
		}

		resp.Licenses = append(resp.Licenses, license)
	}

	return response.JSON(w, resp)
}
