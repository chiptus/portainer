package registry

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "registries"
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

// Registry returns an registry by ID.
func (service *Service) Registry(ID portaineree.RegistryID) (*portaineree.Registry, error) {
	var registry portaineree.Registry
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &registry)
	if err != nil {
		return nil, err
	}

	return &registry, nil
}

// Registries returns an array containing all the registries.
func (service *Service) Registries() ([]portaineree.Registry, error) {
	var registries = make([]portaineree.Registry, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Registry{},
		func(obj interface{}) (interface{}, error) {
			registry, ok := obj.(*portaineree.Registry)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to Registry object")
				return nil, fmt.Errorf("Failed to convert to Registry object: %s", obj)
			}
			registries = append(registries, *registry)
			return &portaineree.Registry{}, nil
		})

	return registries, err
}

// CreateRegistry creates a new registry.
func (service *Service) Create(registry *portaineree.Registry) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			registry.ID = portaineree.RegistryID(id)
			return int(registry.ID), registry
		},
	)
}

// UpdateRegistry updates an registry.
func (service *Service) UpdateRegistry(ID portaineree.RegistryID, registry *portaineree.Registry) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, registry)
}

// DeleteRegistry deletes an registry.
func (service *Service) DeleteRegistry(ID portaineree.RegistryID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}
