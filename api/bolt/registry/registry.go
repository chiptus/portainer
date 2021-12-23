package registry

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"

	"github.com/boltdb/bolt"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "registries"
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

// Registry returns an registry by ID.
func (service *Service) Registry(ID portaineree.RegistryID) (*portaineree.Registry, error) {
	var registry portaineree.Registry
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &registry)
	if err != nil {
		return nil, err
	}

	return &registry, nil
}

// Registries returns an array containing all the registries.
func (service *Service) Registries() ([]portaineree.Registry, error) {
	var registries = make([]portaineree.Registry, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var registry portaineree.Registry
			err := internal.UnmarshalObject(v, &registry)
			if err != nil {
				return err
			}
			registries = append(registries, registry)
		}

		return nil
	})

	return registries, err
}

// CreateRegistry creates a new registry.
func (service *Service) CreateRegistry(registry *portaineree.Registry) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		id, _ := bucket.NextSequence()
		registry.ID = portaineree.RegistryID(id)

		data, err := internal.MarshalObject(registry)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(registry.ID)), data)
	})
}

// UpdateRegistry updates an registry.
func (service *Service) UpdateRegistry(ID portaineree.RegistryID, registry *portaineree.Registry) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, registry)
}

// DeleteRegistry deletes an registry.
func (service *Service) DeleteRegistry(ID portaineree.RegistryID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}
