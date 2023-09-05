package backup

import (
	"net/http"

	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

// @id BackupSettingsFetch
// @summary Fetch s3 backup settings/configurations
// @description **Access policy**: administrator
// @tags backup
// @security ApiKeyAuth
// @security jwt
// @produce json
// @success 200 {object} portaineree.S3BackupSettings "Success"
// @failure 401 "Unauthorized"
// @failure 500 "Server error"
// @router /backup/s3/settings [get]
func (h *Handler) backupSettingsFetch(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	settings, err := h.dataStore.S3Backup().GetSettings()
	if err != nil {
		return httperror.InternalServerError("Unable to retrieve backup settings from the database", err)
	}
	return response.JSON(w, settings)
}
