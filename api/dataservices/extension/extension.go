package extension

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "extension"
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

// Extension returns a extension by ID
func (service *Service) Extension(ID portaineree.ExtensionID) (*portaineree.Extension, error) {
	var extension portaineree.Extension
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &extension)
	if err != nil {
		return nil, err
	}

	return &extension, nil
}

// Extensions return an array containing all the extensions.
func (service *Service) Extensions() ([]portaineree.Extension, error) {
	var extensions = make([]portaineree.Extension, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Extension{},
		func(obj interface{}) (interface{}, error) {
			extension, ok := obj.(*portaineree.Extension)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to Extension object")
				return nil, fmt.Errorf("Failed to convert to Extension object: %s", obj)
			}

			extensions = append(extensions, *extension)

			return &portaineree.Extension{}, nil
		})

	return extensions, err
}

// Persist persists a extension inside the database.
func (service *Service) Persist(extension *portaineree.Extension) error {
	return service.connection.CreateObjectWithId(BucketName, int(extension.ID), extension)
}

// DeleteExtension deletes a Extension.
func (service *Service) DeleteExtension(ID portaineree.ExtensionID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}
