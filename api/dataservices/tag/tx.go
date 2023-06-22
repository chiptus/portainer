package tag

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.Tag, portaineree.TagID]
}

// CreateTag creates a new tag.
func (service ServiceTx) Create(tag *portaineree.Tag) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			tag.ID = portaineree.TagID(id)
			return int(tag.ID), tag
		},
	)
}

// UpdateTagFunc is a no-op inside a transaction.
func (service ServiceTx) UpdateTagFunc(ID portaineree.TagID, updateFunc func(tag *portaineree.Tag)) error {
	return errors.New("cannot be called inside a transaction")
}
