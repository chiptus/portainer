package ssl

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "ssl"
	key        = "SSL"
)

// Service represents a service for managing ssl data.
type Service struct {
	connection *internal.DbConnection
}

// NewService creates a new instance of a service.
func NewService(connection *internal.DbConnection) (*Service, error) {
	err := internal.CreateBucket(connection, BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// Settings retrieve the ssl settings object.
func (service *Service) Settings() (*portaineree.SSLSettings, error) {
	var settings portaineree.SSLSettings

	err := internal.GetObject(service.connection, BucketName, []byte(key), &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// UpdateSettings persists a SSLSettings object.
func (service *Service) UpdateSettings(settings *portaineree.SSLSettings) error {
	return internal.UpdateObject(service.connection, BucketName, []byte(key), settings)
}
