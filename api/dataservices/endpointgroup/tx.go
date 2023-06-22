package endpointgroup

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.EndpointGroup, portaineree.EndpointGroupID]
}

// CreateEndpointGroup assign an ID to a new environment(endpoint) group and saves it.
func (service ServiceTx) Create(endpointGroup *portaineree.EndpointGroup) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			endpointGroup.ID = portaineree.EndpointGroupID(id)
			return int(endpointGroup.ID), endpointGroup
		},
	)
}
