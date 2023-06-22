package registry

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.Registry, portaineree.RegistryID]
}

// Create creates a new registry.
func (service ServiceTx) Create(registry *portaineree.Registry) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			registry.ID = portaineree.RegistryID(id)
			return int(registry.ID), registry
		},
	)
}
