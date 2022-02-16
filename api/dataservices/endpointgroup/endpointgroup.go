package endpointgroup

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "endpoint_groups"
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

// EndpointGroup returns an environment(endpoint) group by ID.
func (service *Service) EndpointGroup(ID portaineree.EndpointGroupID) (*portaineree.EndpointGroup, error) {
	var endpointGroup portaineree.EndpointGroup
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &endpointGroup)
	if err != nil {
		return nil, err
	}

	return &endpointGroup, nil
}

// UpdateEndpointGroup updates an environment(endpoint) group.
func (service *Service) UpdateEndpointGroup(ID portaineree.EndpointGroupID, endpointGroup *portaineree.EndpointGroup) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, endpointGroup)
}

// DeleteEndpointGroup deletes an environment(endpoint) group.
func (service *Service) DeleteEndpointGroup(ID portaineree.EndpointGroupID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// EndpointGroups return an array containing all the environment(endpoint) groups.
func (service *Service) EndpointGroups() ([]portaineree.EndpointGroup, error) {
	var endpointGroups = make([]portaineree.EndpointGroup, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.EndpointGroup{},
		func(obj interface{}) (interface{}, error) {
			//var tag portaineree.Tag
			endpointGroup, ok := obj.(*portaineree.EndpointGroup)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to EndpointGroup object")
				return nil, fmt.Errorf("Failed to convert to EndpointGroup object: %s", obj)
			}
			endpointGroups = append(endpointGroups, *endpointGroup)
			return &portaineree.EndpointGroup{}, nil
		})

	return endpointGroups, err
}

// CreateEndpointGroup assign an ID to a new environment(endpoint) group and saves it.
func (service *Service) Create(endpointGroup *portaineree.EndpointGroup) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			endpointGroup.ID = portaineree.EndpointGroupID(id)
			return int(endpointGroup.ID), endpointGroup
		},
	)
}
