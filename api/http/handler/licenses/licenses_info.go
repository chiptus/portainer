package licenses

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id licensesInfo
// @summary summarizes licenses on Portainer
// @description
// @description **Access policy**: administrator
// @tags license
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {object} portaineree.LicenseInfo "License info"
// @router /licenses/info [get]
func (handler *Handler) licensesInfo(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	info := handler.LicenseService.Info()

	return response.JSON(w, info)
}
