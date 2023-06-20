package role

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

// Role returns a Role by ID
func (service ServiceTx) Role(ID portaineree.RoleID) (*portaineree.Role, error) {
	var set portaineree.Role
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &set)
	if err != nil {
		return nil, err
	}

	return &set, nil
}

// Roles returns an array containing all the sets.
func (service ServiceTx) Roles() ([]portaineree.Role, error) {
	var sets = make([]portaineree.Role, 0)

	return sets, service.tx.GetAll(
		BucketName,
		&portaineree.Role{},
		dataservices.AppendFn(&sets),
	)
}

// CreateRole creates a new Role.
func (service ServiceTx) Create(role *portaineree.Role) error {
	return service.tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			role.ID = portaineree.RoleID(id)
			return int(role.ID), role
		},
	)
}

// UpdateRole updates a role.
func (service ServiceTx) UpdateRole(ID portaineree.RoleID, role *portaineree.Role) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, role)
}
