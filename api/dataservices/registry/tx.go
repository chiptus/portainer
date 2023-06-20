package registry

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
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

	return registries, service.tx.GetAll(
		BucketName,
		&portaineree.Registry{},
		dataservices.AppendFn(&registries),
	)
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
