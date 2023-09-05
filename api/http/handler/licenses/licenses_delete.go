package licenses

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type (
	deletePayload struct {
		// List of license keys to remove
		LicenseKeys []string
	}
)

func (payload *deletePayload) Validate(r *http.Request) error {
	if len(payload.LicenseKeys) == 0 {
		return errors.New("missing licenses keys")
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
// @success 200
// @router /licenses/remove [post]
func (handler *Handler) licensesDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload deletePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	for _, licenseKey := range payload.LicenseKeys {
		err := handler.LicenseService.DeleteLicense(licenseKey)
		if err != nil {
			return httperror.InternalServerError("Failed deleting license(s)", err)
		}
	}

	return response.Empty(w)
}
