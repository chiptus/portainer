package customtemplate

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "customtemplates"

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

	return customTemplates, service.connection.GetAll(
		BucketName,
		&portaineree.CustomTemplate{},
		dataservices.AppendFn(&customTemplates),
	)
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
