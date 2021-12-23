package settings

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName  = "settings"
	settingsKey = "SETTINGS"
)

// Service represents a service for managing environment(endpoint) data.
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

// Settings retrieve the settings object.
func (service *Service) Settings() (*portaineree.Settings, error) {
	var settings portaineree.Settings

	err := internal.GetObject(service.connection, BucketName, []byte(settingsKey), &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// UpdateSettings persists a Settings object.
func (service *Service) UpdateSettings(settings *portaineree.Settings) error {
	return internal.UpdateObject(service.connection, BucketName, []byte(settingsKey), settings)
}

func (service *Service) IsFeatureFlagEnabled(feature portaineree.Feature) bool {
	settings, err := service.Settings()
	if err != nil {
		return false
	}

	featureFlagSetting, ok := settings.FeatureFlagSettings[feature]
	if ok {
		return featureFlagSetting
	}

	return false
}
