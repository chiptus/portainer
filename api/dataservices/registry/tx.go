package registry

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

// Registry returns a registry by ID.
func (service ServiceTx) Registry(ID portaineree.RegistryID) (*portaineree.Registry, error) {
	var registry portaineree.Registry
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &registry)
	if err != nil {
		return nil, err
	}

	return &registry, nil
}

// Registries returns an array containing all the registries.
func (service ServiceTx) Registries() ([]portaineree.Registry, error) {
	var registries = make([]portaineree.Registry, 0)

	err := service.tx.GetAll(
		BucketName,
		&portaineree.Registry{},
		func(obj interface{}) (interface{}, error) {
			registry, ok := obj.(*portaineree.Registry)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to Registry object")
				return nil, fmt.Errorf("Failed to convert to Registry object: %s", obj)
			}

			registries = append(registries, *registry)

			return &portaineree.Registry{}, nil
		})

	return registries, err
}

// Create creates a new registry.
func (service ServiceTx) Create(registry *portaineree.Registry) error {
	return service.tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			registry.ID = portaineree.RegistryID(id)
			return int(registry.ID), registry
		},
	)
}

// UpdateRegistry updates a registry.
func (service ServiceTx) UpdateRegistry(ID portaineree.RegistryID, registry *portaineree.Registry) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, registry)
}

// DeleteRegistry deletes a registry.
func (service ServiceTx) DeleteRegistry(ID portaineree.RegistryID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}
