package tag

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"

	"github.com/boltdb/bolt"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "tags"
)

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection *internal.DbConnection
}

// NewService creates a new instance of a service.
func NewService(connection *internal.DbConnection) (*Service, error) {
	err := internal.CreateBucket(connection, BucketName)
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

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var tag portaineree.Tag
			err := internal.UnmarshalObject(v, &tag)
			if err != nil {
				return err
			}
			tags = append(tags, tag)
		}

		return nil
	})

	return tags, err
}

// Tag returns a tag by ID.
func (service *Service) Tag(ID portaineree.TagID) (*portaineree.Tag, error) {
	var tag portaineree.Tag
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &tag)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

// CreateTag creates a new tag.
func (service *Service) CreateTag(tag *portaineree.Tag) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		id, _ := bucket.NextSequence()
		tag.ID = portaineree.TagID(id)

		data, err := internal.MarshalObject(tag)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(tag.ID)), data)
	})
}

// UpdateTag updates a tag.
func (service *Service) UpdateTag(ID portaineree.TagID, tag *portaineree.Tag) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, tag)
}

// DeleteTag deletes a tag.
func (service *Service) DeleteTag(ID portaineree.TagID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}
