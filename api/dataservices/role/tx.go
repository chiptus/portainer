package role

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.Role, portainer.RoleID]
}

// CreateRole creates a new Role.
func (service ServiceTx) Create(role *portaineree.Role) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			role.ID = portainer.RoleID(id)
			return int(role.ID), role
		},
	)
}
