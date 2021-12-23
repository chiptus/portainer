package backup

import (
	"net/http"

	"github.com/robfig/cron/v3"

	"github.com/pkg/errors"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
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
	if p.Region == "" {
		return errors.New("missing Region")
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
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	if err := h.backupScheduler.Update(payload.S3BackupSettings); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Couldn't update backup settings", Err: err}
	}

	return nil
}
