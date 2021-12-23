package customtemplate

import (
	"github.com/boltdb/bolt"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "customtemplates"
)

// Service represents a service for managing custom template data.
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

// CustomTemplates return an array containing all the custom templates.
func (service *Service) CustomTemplates() ([]portaineree.CustomTemplate, error) {
	var customTemplates = make([]portaineree.CustomTemplate, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var customTemplate portaineree.CustomTemplate
			err := internal.UnmarshalObjectWithJsoniter(v, &customTemplate)
			if err != nil {
				return err
			}
			customTemplates = append(customTemplates, customTemplate)
		}

		return nil
	})

	return customTemplates, err
}

// CustomTemplate returns an custom template by ID.
func (service *Service) CustomTemplate(ID portaineree.CustomTemplateID) (*portaineree.CustomTemplate, error) {
	var customTemplate portaineree.CustomTemplate
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &customTemplate)
	if err != nil {
		return nil, err
	}

	return &customTemplate, nil
}

// UpdateCustomTemplate updates an custom template.
func (service *Service) UpdateCustomTemplate(ID portaineree.CustomTemplateID, customTemplate *portaineree.CustomTemplate) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, customTemplate)
}

// DeleteCustomTemplate deletes an custom template.
func (service *Service) DeleteCustomTemplate(ID portaineree.CustomTemplateID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}

// CreateCustomTemplate assign an ID to a new custom template and saves it.
func (service *Service) CreateCustomTemplate(customTemplate *portaineree.CustomTemplate) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		data, err := internal.MarshalObject(customTemplate)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(customTemplate.ID)), data)
	})
}

// GetNextIdentifier returns the next identifier for a custom template.
func (service *Service) GetNextIdentifier() int {
	return internal.GetNextIdentifier(service.connection, BucketName)
}
