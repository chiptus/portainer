package endpointrelation

import (
	"github.com/boltdb/bolt"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "endpoint_relations"
)

// Service represents a service for managing environment(endpoint) relation data.
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

// EndpointRelation returns a Environment(Endpoint) relation object by EndpointID
func (service *Service) EndpointRelation(endpointID portaineree.EndpointID) (*portaineree.EndpointRelation, error) {
	var endpointRelation portaineree.EndpointRelation
	identifier := internal.Itob(int(endpointID))

	err := internal.GetObject(service.connection, BucketName, identifier, &endpointRelation)
	if err != nil {
		return nil, err
	}

	return &endpointRelation, nil
}

// CreateEndpointRelation saves endpointRelation
func (service *Service) CreateEndpointRelation(endpointRelation *portaineree.EndpointRelation) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		data, err := internal.MarshalObject(endpointRelation)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(endpointRelation.EndpointID)), data)
	})
}

// UpdateEndpointRelation updates an Environment(Endpoint) relation object
func (service *Service) UpdateEndpointRelation(EndpointID portaineree.EndpointID, endpointRelation *portaineree.EndpointRelation) error {
	identifier := internal.Itob(int(EndpointID))
	return internal.UpdateObject(service.connection, BucketName, identifier, endpointRelation)
}

// DeleteEndpointRelation deletes an Environment(Endpoint) relation object
func (service *Service) DeleteEndpointRelation(EndpointID portaineree.EndpointID) error {
	identifier := internal.Itob(int(EndpointID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}
