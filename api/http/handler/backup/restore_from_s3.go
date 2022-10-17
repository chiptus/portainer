package backup

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portaineree "github.com/portainer/portainer-ee/api"
	operations "github.com/portainer/portainer-ee/api/backup"
	s3client "github.com/portainer/portainer-ee/api/s3"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type restoreS3Settings struct {
	portaineree.S3Location
	Password string
}

func (p *restoreS3Settings) Validate(r *http.Request) error {
	if p.AccessKeyID == "" {
		return errors.New("missing AccessKeyID field")
	}
	if p.SecretAccessKey == "" {
		return errors.New("missing SecretAccessKe field")
	}
	if p.Region == "" {
		return errors.New("missing Region field")
	}
	if p.BucketName == "" {
		return errors.New("missing BucketName field")
	}
	if p.Filename == "" {
		return errors.New("missing Filename field")
	}
	return nil
}

// @id RestoreFromS3
// @summary Triggers a system restore using details of s3 backup
// @description Triggers a system restore using details of s3 backup
// @description **Access policy**: public
// @tags backup
// @accept json
// @param body body restoreS3Settings false "S3 Location Payload"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /backup/s3/restore [post]
func (h *Handler) restoreFromS3(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	initialized, err := h.adminMonitor.WasInitialized()
	if err != nil {
		return httperror.InternalServerError("Failed to check system initialization", err)
	}
	if initialized {
		return httperror.BadRequest("Cannot restore already initialized instance", fmt.Errorf("system already initialized"))
	}

	h.adminMonitor.Stop()
	defer h.adminMonitor.Start()

	var payload restoreS3Settings
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	backupFile, err := createTmpBackupLocation(h.filestorePath)
	if err != nil {
		log.Debug().Err(err).Msg("")

		return httperror.InternalServerError("Failed to restore", err)
	}

	defer func() {
		backupFile.Close()
		os.RemoveAll(filepath.Dir(backupFile.Name()))
	}()

	s3session, err := s3client.NewSession(payload.Region, payload.AccessKeyID, payload.SecretAccessKey)
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	if err = s3client.Download(s3session, backupFile, payload.S3Location); err != nil {
		log.Error().Err(err).Msg("failed downloading file from S3")

		return httperror.InternalServerError("Failed to download file from S3", err)
	}

	if err = operations.RestoreArchive(backupFile, payload.Password, h.filestorePath, h.gate, h.dataStore, h.userActivityStore, h.shutdownTrigger); err != nil {
		log.Error().Err(err).Msg("failed to restore system from backup")

		return httperror.InternalServerError("Failed to restore backup", err)
	}

	return nil
}

func createTmpBackupLocation(filestorePath string) (*os.File, error) {
	restoreDir, err := os.MkdirTemp(filestorePath, fmt.Sprintf("restore_%s", time.Now().Format("2006-01-02_15-04-05")))
	if err != nil {
		return nil, errors.New("failed to create tmp download dir")
	}

	f, err := os.Create(filepath.Join(restoreDir, "backup_file"))
	if err != nil {
		return nil, errors.New("failed to create tmp download file")
	}

	return f, nil
}
