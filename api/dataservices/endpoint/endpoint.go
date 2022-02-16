package endpoint

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "endpoints"
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

// Endpoint returns an environment(endpoint) by ID.
func (service *Service) Endpoint(ID portaineree.EndpointID) (*portaineree.Endpoint, error) {
	var endpoint portaineree.Endpoint
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &endpoint)
	if err != nil {
		return nil, err
	}

	return &endpoint, nil
}

// UpdateEndpoint updates an environment(endpoint).
func (service *Service) UpdateEndpoint(ID portaineree.EndpointID, endpoint *portaineree.Endpoint) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, endpoint)
}

// DeleteEndpoint deletes an environment(endpoint).
func (service *Service) DeleteEndpoint(ID portaineree.EndpointID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// Endpoints return an array containing all the environments(endpoints).
func (service *Service) Endpoints() ([]portaineree.Endpoint, error) {
	var endpoints = make([]portaineree.Endpoint, 0)

	err := service.connection.GetAllWithJsoniter(
		BucketName,
		&portaineree.Endpoint{},
		func(obj interface{}) (interface{}, error) {
			endpoint, ok := obj.(*portaineree.Endpoint)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to Endpoint object")
				return nil, fmt.Errorf("failed to convert to Endpoint object: %s", obj)
			}
			endpoints = append(endpoints, *endpoint)
			return &portaineree.Endpoint{}, nil
		})

	return endpoints, err
}

// CreateEndpoint assign an ID to a new environment(endpoint) and saves it.
func (service *Service) Create(endpoint *portaineree.Endpoint) error {
	return service.connection.CreateObjectWithSetSequence(BucketName, int(endpoint.ID), endpoint)
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
