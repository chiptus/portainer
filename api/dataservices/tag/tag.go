package tag

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "tags"
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

// Tags return an array containing all the tags.
func (service *Service) Tags() ([]portaineree.Tag, error) {
	var tags = make([]portaineree.Tag, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.Tag{},
		func(obj interface{}) (interface{}, error) {
			tag, ok := obj.(*portaineree.Tag)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to Tag object")
				return nil, fmt.Errorf("Failed to convert to Tag object: %s", obj)
			}

			tags = append(tags, *tag)

			return &portaineree.Tag{}, nil
		})

	return tags, err
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

// UpdateTag updates a tag.
func (service *Service) UpdateTag(ID portaineree.TagID, tag *portaineree.Tag) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, tag)
}

// DeleteTag deletes a tag.
func (service *Service) DeleteTag(ID portaineree.TagID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}
