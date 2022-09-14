package licenses

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id licensesList
// @summary fetches the list of licenses on Portainer
// @description
// @description **Access policy**: administrator
// @tags license
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {array} liblicense.PortainerLicense "Licenses"
// @router /licenses [get]
func (handler *Handler) licensesList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	licenses, err := handler.LicenseService.Licenses()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve Licenses from the database", err)
	}

	if handler.demoService.IsDemo() {
		for i := range licenses {
			licenses[i].LicenseKey = "This feature is not available in the demo version of Portainer"
		}
	}

	return response.JSON(w, licenses)
}
