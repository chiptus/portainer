package role

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "roles"
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

// Role returns a Role by ID
func (service *Service) Role(ID portaineree.RoleID) (*portaineree.Role, error) {
	var set portaineree.Role
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &set)
	if err != nil {
		return nil, err
	}

	return &set, nil
}

// Roles return an array containing all the sets.
func (service *Service) Roles() ([]portaineree.Role, error) {
	var sets = make([]portaineree.Role, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Role{},
		func(obj interface{}) (interface{}, error) {
			set, ok := obj.(*portaineree.Role)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to Role object")
				return nil, fmt.Errorf("Failed to convert to Role object: %s", obj)
			}
			sets = append(sets, *set)
			return &portaineree.Role{}, nil
		})

	return sets, err
}

// CreateRole creates a new Role.
func (service *Service) Create(role *portaineree.Role) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			role.ID = portaineree.RoleID(id)
			return int(role.ID), role
		},
	)
}

// UpdateRole updates a role.
func (service *Service) UpdateRole(ID portaineree.RoleID, role *portaineree.Role) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, role)
}
