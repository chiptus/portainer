package edgegroup

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "edgegroups"
)

// Service represents a service for managing Edge group data.
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

// EdgeGroups return an array containing all the Edge groups.
func (service *Service) EdgeGroups() ([]portaineree.EdgeGroup, error) {
	var groups = make([]portaineree.EdgeGroup, 0)

	err := service.connection.GetAllWithJsoniter(
		BucketName,
		&portaineree.EdgeGroup{},
		func(obj interface{}) (interface{}, error) {
			group, ok := obj.(*portaineree.EdgeGroup)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to EdgeGroup object")
				return nil, fmt.Errorf("Failed to convert to EdgeGroup object: %s", obj)
			}
			groups = append(groups, *group)
			return &portaineree.EdgeGroup{}, nil
		})

	return groups, err
}

// EdgeGroup returns an Edge group by ID.
func (service *Service) EdgeGroup(ID portaineree.EdgeGroupID) (*portaineree.EdgeGroup, error) {
	var group portaineree.EdgeGroup
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

// UpdateEdgeGroup updates an Edge group.
func (service *Service) UpdateEdgeGroup(ID portaineree.EdgeGroupID, group *portaineree.EdgeGroup) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, group)
}

// DeleteEdgeGroup deletes an Edge group.
func (service *Service) DeleteEdgeGroup(ID portaineree.EdgeGroupID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// CreateEdgeGroup assign an ID to a new Edge group and saves it.
func (service *Service) Create(group *portaineree.EdgeGroup) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			group.ID = portaineree.EdgeGroupID(id)
			return int(group.ID), group
		},
	)
}
