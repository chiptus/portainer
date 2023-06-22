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
	dataservices.BaseDataService[portaineree.Tag, portaineree.TagID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.Tag, portaineree.TagID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.Tag, portaineree.TagID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// CreateTag creates a new tag.
func (service *Service) Create(tag *portaineree.Tag) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			tag.ID = portaineree.TagID(id)
			return int(tag.ID), tag
		},
	)
}

// UpdateTagFunc updates a tag inside a transaction avoiding data races.
func (service *Service) UpdateTagFunc(ID portaineree.TagID, updateFunc func(tag *portaineree.Tag)) error {
	id := service.Connection.ConvertToKey(int(ID))
	tag := &portaineree.Tag{}

	return service.Connection.UpdateObjectFunc(BucketName, id, tag, func() {
		updateFunc(tag)
	})
}
