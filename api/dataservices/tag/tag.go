package tag

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "tags"

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

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		service: service,
		tx:      tx,
	}
}

// Tags return an array containing all the tags.
func (service *Service) Tags() ([]portaineree.Tag, error) {
	var tags = make([]portaineree.Tag, 0)

	return tags, service.connection.GetAll(
		BucketName,
		&portaineree.Tag{},
		dataservices.AppendFn(&tags),
	)
}

// Tag returns a tag by ID.
func (service *Service) Tag(ID portaineree.TagID) (*portaineree.Tag, error) {
	var tag portaineree.Tag
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &tag)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

// CreateTag creates a new tag.
func (service *Service) Create(tag *portaineree.Tag) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			tag.ID = portaineree.TagID(id)
			return int(tag.ID), tag
		},
	)
}

// Deprecated: Use UpdateTagFunc instead.
func (service *Service) UpdateTag(ID portaineree.TagID, tag *portaineree.Tag) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, tag)
}

// UpdateTagFunc updates a tag inside a transaction avoiding data races.
func (service *Service) UpdateTagFunc(ID portaineree.TagID, updateFunc func(tag *portaineree.Tag)) error {
	id := service.connection.ConvertToKey(int(ID))
	tag := &portaineree.Tag{}

	return service.connection.UpdateObjectFunc(BucketName, id, tag, func() {
		updateFunc(tag)
	})
}

// DeleteTag deletes a tag.
func (service *Service) DeleteTag(ID portaineree.TagID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}
