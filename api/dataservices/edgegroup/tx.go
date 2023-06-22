package edgegroup

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.EdgeGroup, portaineree.EdgeGroupID]
}

// UpdateEdgeGroupFunc is a no-op inside a transaction.
func (service ServiceTx) UpdateEdgeGroupFunc(ID portaineree.EdgeGroupID, updateFunc func(edgeGroup *portaineree.EdgeGroup)) error {
	return errors.New("cannot be called inside a transaction")
}

func (service ServiceTx) Create(group *portaineree.EdgeGroup) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			group.ID = portaineree.EdgeGroupID(id)
			return int(group.ID), group
		},
	)
}
