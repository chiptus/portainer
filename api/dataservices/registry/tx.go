package registry

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.Registry, portainer.RegistryID]
}

// Create creates a new registry.
func (service ServiceTx) Create(registry *portaineree.Registry) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			registry.ID = portainer.RegistryID(id)
			return int(registry.ID), registry
		},
	)
}
