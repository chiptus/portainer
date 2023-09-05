package backup

import (
	"net/http"

	portaineree "github.com/portainer/portainer-ee/api"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

type backupSettings struct {
	portaineree.S3BackupSettings
}

func (p *backupSettings) Validate(r *http.Request) error {
	if p.CronRule == "" {
		return nil
	}
	if _, err := cron.ParseStandard(p.CronRule); err != nil {
		return errors.New("invalid cron rule")
	}
	if p.AccessKeyID == "" {
		return errors.New("missing AccessKeyID")
	}
	if p.SecretAccessKey == "" {
		return errors.New("missing SecretAccessKey")
	}
	if p.BucketName == "" {
		return errors.New("missing BucketName")
	}
	return nil
}

// @id UpdateS3Settings
// @summary Updates stored s3 backup settings and updates running cron jobs as needed
// @description Updates stored s3 backup settings and updates running cron jobs as needed
// @description **Access policy**: administrator
// @tags backup
// @security ApiKeyAuth
// @security jwt
// @accept json
// @param s3_backup_settings body portaineree.S3BackupSettings false "S3 backup settings"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /backup/s3/settings [post]
func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload backupSettings
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	if err := h.backupScheduler.Update(payload.S3BackupSettings); err != nil {
		return httperror.InternalServerError("Couldn't update backup settings", err)
	}

	return nil
}
