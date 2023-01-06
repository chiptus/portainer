package endpointrelation

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/edge/cache"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "endpoint_relations"
)

// Service represents a service for managing environment(endpoint) relation data.
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

// EndpointRelations returns an array of all EndpointRelations
func (service *Service) EndpointRelations() ([]portaineree.EndpointRelation, error) {
	var all = make([]portaineree.EndpointRelation, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.EndpointRelation{},
		func(obj interface{}) (interface{}, error) {
			r, ok := obj.(*portaineree.EndpointRelation)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to EndpointRelation object")
				return nil, fmt.Errorf("Failed to convert to EndpointRelation object: %s", obj)
			}

			all = append(all, *r)

			return &portaineree.EndpointRelation{}, nil
		})

	return all, err
}

// EndpointRelation returns a Environment(Endpoint) relation object by EndpointID
func (service *Service) EndpointRelation(endpointID portaineree.EndpointID) (*portaineree.EndpointRelation, error) {
	var endpointRelation portaineree.EndpointRelation
	identifier := service.connection.ConvertToKey(int(endpointID))

	err := service.connection.GetObject(BucketName, identifier, &endpointRelation)
	if err != nil {
		return nil, err
	}

	return &endpointRelation, nil
}

// CreateEndpointRelation saves endpointRelation
func (service *Service) Create(endpointRelation *portaineree.EndpointRelation) error {
	err := service.connection.CreateObjectWithId(BucketName, int(endpointRelation.EndpointID), endpointRelation)
	cache.Del(endpointRelation.EndpointID)

	return err
}

// UpdateEndpointRelation updates an Environment(Endpoint) relation object
func (service *Service) UpdateEndpointRelation(endpointID portaineree.EndpointID, endpointRelation *portaineree.EndpointRelation) error {
	identifier := service.connection.ConvertToKey(int(endpointID))
	err := service.connection.UpdateObject(BucketName, identifier, endpointRelation)
	cache.Del(endpointID)

	return err
}

// DeleteEndpointRelation deletes an Environment(Endpoint) relation object
func (service *Service) DeleteEndpointRelation(endpointID portaineree.EndpointID) error {
	identifier := service.connection.ConvertToKey(int(endpointID))
	err := service.connection.DeleteObject(BucketName, identifier)
	cache.Del(endpointID)

	return err
}
