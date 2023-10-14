package edgegroup

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.EdgeGroup, portainer.EdgeGroupID]
}

// UpdateEdgeGroupFunc is a no-op inside a transaction.
func (service ServiceTx) UpdateEdgeGroupFunc(ID portainer.EdgeGroupID, updateFunc func(edgeGroup *portaineree.EdgeGroup)) error {
	return errors.New("cannot be called inside a transaction")
}

func (service ServiceTx) Create(group *portaineree.EdgeGroup) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			group.ID = portainer.EdgeGroupID(id)
			return int(group.ID), group
		},
	)
}
