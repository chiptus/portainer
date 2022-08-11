package enforcement

import (
	"github.com/pkg/errors"
	"github.com/portainer/portainer-ee/api/database/models"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/dataservices/errors"
)

const BucketName = "enforcement"

var LicenseOveruseEnforcementKey = []byte("1-TGljZW5zZU92ZXJ1c2VFbmZvcmNlbWVudEtleQ==")

// Service manages license enforcement record
type Service struct {
	connection portainer.Connection
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// LicenseEnforcement returns a license enforcement record
func (s *Service) LicenseEnforcement() (*models.LicenseEnforcement, error) {
	var enforcement models.LicenseEnforcement
	err := s.connection.GetObject(BucketName, LicenseOveruseEnforcementKey, &enforcement)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return nil, errors.Wrapf(err, "failed to fetch the enforcement record with key %s", string(LicenseOveruseEnforcementKey))
	}

	return &enforcement, nil
}

// UpdateOveruseStartedTimestamp sets new overuse start timestamp if changed
func (s *Service) UpdateOveruseStartedTimestamp(timestamp int64) error {
	record, err := s.LicenseEnforcement()
	if err != nil {
		return err
	}

	if record == nil {
		record = &models.LicenseEnforcement{}
	}

	if record.LicenseOveruseStartedTimestamp == timestamp {
		return nil
	}

	record.LicenseOveruseStartedTimestamp = timestamp
	err = s.connection.UpdateObject(BucketName, LicenseOveruseEnforcementKey, record)
	return errors.Wrapf(err, "failed to update the enforcement record with key %s", string(LicenseOveruseEnforcementKey))
}
