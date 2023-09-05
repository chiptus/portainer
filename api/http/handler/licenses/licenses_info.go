package licenses

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type LicenseInfo struct {
	*portaineree.LicenseInfo
	EnforcedAt int64 `json:"enforcedAt"`
}

// @id licensesInfo
// @summary summarizes licenses on Portainer
// @description
// @description **Access policy**: administrator
// @tags license
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {object} LicenseInfo "License info"
// @router /licenses/info [get]
func (handler *Handler) licensesInfo(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	info := handler.LicenseService.Info()

	result := &LicenseInfo{
		LicenseInfo: &info,
		EnforcedAt:  handler.LicenseService.WillBeEnforcedAt(),
	}

	return response.JSON(w, result)
}
