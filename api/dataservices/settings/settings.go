package settings

import (
	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName  = "settings"
	settingsKey = "SETTINGS"
)

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection portainer.Connection
}

func (service *Service) BucketName() string {
	return BucketName
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		service: service,
		tx:      tx,
	}
}

// Settings retrieve the settings object.
func (service *Service) Settings() (*portaineree.Settings, error) {
	var settings portaineree.Settings

	err := service.connection.GetObject(BucketName, []byte(settingsKey), &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// UpdateSettings persists a Settings object.
func (service *Service) UpdateSettings(settings *portaineree.Settings) error {
	return service.connection.UpdateObject(BucketName, []byte(settingsKey), settings)
}
