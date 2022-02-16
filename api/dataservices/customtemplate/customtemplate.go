package customtemplate

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "customtemplates"
)

// Service represents a service for managing custom template data.
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

// CustomTemplates return an array containing all the custom templates.
func (service *Service) CustomTemplates() ([]portaineree.CustomTemplate, error) {
	var customTemplates = make([]portaineree.CustomTemplate, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.CustomTemplate{},
		func(obj interface{}) (interface{}, error) {
			//var tag portainer.Tag
			customTemplate, ok := obj.(*portaineree.CustomTemplate)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to CustomTemplate object")
				return nil, fmt.Errorf("Failed to convert to CustomTemplate object: %s", obj)
			}
			customTemplates = append(customTemplates, *customTemplate)
			return &portaineree.CustomTemplate{}, nil
		})

	return customTemplates, err
}

// CustomTemplate returns an custom template by ID.
func (service *Service) CustomTemplate(ID portaineree.CustomTemplateID) (*portaineree.CustomTemplate, error) {
	var customTemplate portaineree.CustomTemplate
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &customTemplate)
	if err != nil {
		return nil, err
	}

	return &customTemplate, nil
}

// UpdateCustomTemplate updates an custom template.
func (service *Service) UpdateCustomTemplate(ID portaineree.CustomTemplateID, customTemplate *portaineree.CustomTemplate) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, customTemplate)
}

// DeleteCustomTemplate deletes an custom template.
func (service *Service) DeleteCustomTemplate(ID portaineree.CustomTemplateID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// CreateCustomTemplate uses the existing id and saves it.
// TODO: where does the ID come from, and is it safe?
func (service *Service) Create(customTemplate *portaineree.CustomTemplate) error {
	return service.connection.CreateObjectWithId(BucketName, int(customTemplate.ID), customTemplate)
}

// GetNextIdentifier returns the next identifier for a custom template.
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
