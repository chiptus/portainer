package s3backup

import (
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	perrors "github.com/portainer/portainer/api/dataservices/errors"
)

const (
	bucketName  = "s3backup"
	statusKey   = "lastRunStatus"
	settingsKey = "settings"
)

type Service struct {
	connection portainer.Connection
}

// NewService creates a new service and ensures corresponding bucket exist
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(bucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// GetStatus returns the status of the last scheduled backup run
func (s *Service) GetStatus() (portaineree.S3BackupStatus, error) {
	var status portaineree.S3BackupStatus
	err := s.connection.GetObject(bucketName, []byte(statusKey), &status)
	if errors.Is(err, perrors.ErrObjectNotFound) {
		return status, nil
	}

	return status, err
}

// DropStatus deletes the status of the last sheduled backup run
func (s *Service) DropStatus() error {
	return s.connection.DeleteObject(bucketName, []byte(statusKey))
}

// UpdateStatus upserts a status of the last scheduled backup run
func (s *Service) UpdateStatus(status portaineree.S3BackupStatus) error {
	return s.connection.UpdateObject(bucketName, []byte(statusKey), status)
}

// UpdateSettings updates stored s3 backup settings
func (s *Service) UpdateSettings(settings portaineree.S3BackupSettings) error {
	return s.connection.UpdateObject(bucketName, []byte(settingsKey), settings)
}

// GetSettings returns stored s3 backup settings
func (s *Service) GetSettings() (portaineree.S3BackupSettings, error) {
	var settings portaineree.S3BackupSettings
	err := s.connection.GetObject(bucketName, []byte(settingsKey), &settings)
	if errors.Is(err, perrors.ErrObjectNotFound) {
		return settings, nil
	}

	return settings, err
}
