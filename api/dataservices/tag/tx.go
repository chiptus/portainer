package tag

import (
	"errors"

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

// Tags return an array containing all the tags.
func (service ServiceTx) Tags() ([]portaineree.Tag, error) {
	var tags = make([]portaineree.Tag, 0)

	return tags, service.tx.GetAll(
		BucketName,
		&portaineree.Tag{},
		dataservices.AppendFn(&tags),
	)
}

// Tag returns a tag by ID.
func (service ServiceTx) Tag(ID portaineree.TagID) (*portaineree.Tag, error) {
	var tag portaineree.Tag
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &tag)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

// CreateTag creates a new tag.
func (service ServiceTx) Create(tag *portaineree.Tag) error {
	return service.tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			tag.ID = portaineree.TagID(id)
			return int(tag.ID), tag
		},
	)
}

// UpdateTag updates a tag.
func (service ServiceTx) UpdateTag(ID portaineree.TagID, tag *portaineree.Tag) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, tag)
}

// UpdateTagFunc is a no-op inside a transaction.
func (service ServiceTx) UpdateTagFunc(ID portaineree.TagID, updateFunc func(tag *portaineree.Tag)) error {
	return errors.New("cannot be called inside a transaction")
}

// DeleteTag deletes a tag.
func (service ServiceTx) DeleteTag(ID portaineree.TagID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}
