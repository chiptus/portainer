package licenses

import (
	"net/http"

	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
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
	licenses := handler.LicenseService.Licenses()
	if handler.demoService.IsDemo() {
		for i := range licenses {
			licenses[i].LicenseKey = "This feature is not available in the demo version of Portainer"
		}
	}

	return response.JSON(w, licenses)
}
