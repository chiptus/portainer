package registry

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "registries"

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	dataservices.BaseDataService[portaineree.Registry, portaineree.RegistryID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.Registry, portaineree.RegistryID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.Registry, portaineree.RegistryID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// Create creates a new registry.
func (service *Service) Create(registry *portaineree.Registry) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			registry.ID = portaineree.RegistryID(id)
			return int(registry.ID), registry
		},
	)
}
