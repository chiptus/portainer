package cloudcredential

import (
	"fmt"
	"time"

	"github.com/portainer/portainer-ee/api/database/models"
	portainer "github.com/portainer/portainer/api"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "cloudcredentials"
)

// Service represents a service for managing cloudcredential data.
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

// GetByID returns an cloudcredential by ID.
func (service *Service) GetByID(ID models.CloudCredentialID) (*models.CloudCredential, error) {
	var cloudcredential models.CloudCredential
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &cloudcredential)
	if err != nil {
		return nil, err
	}

	return &cloudcredential, nil
}

// Upadte updates an cloudcredential.
func (service *Service) Update(ID models.CloudCredentialID, cloudcredential *models.CloudCredential) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, cloudcredential)
}

// Delete deletes an cloudcredential.
func (service *Service) Delete(ID models.CloudCredentialID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// GetAll return an array containing all the cloudcredentials.
func (service *Service) GetAll() ([]models.CloudCredential, error) {
	var cloudcreds = make([]models.CloudCredential, 0)

	err := service.connection.GetAllWithJsoniter(
		BucketName,
		&models.CloudCredential{},
		func(obj interface{}) (interface{}, error) {
			cloudcredential, ok := obj.(*models.CloudCredential)
			if !ok {
				return nil, fmt.Errorf("failed to convert to CloudCredential object: %s", obj)
			}
			cloudcreds = append(cloudcreds, *cloudcredential)
			return &models.CloudCredential{}, nil
		})

	return cloudcreds, err
}

// Create assign an ID to a new cloudcredential and saves it.
func (service *Service) Create(cloudcredential *models.CloudCredential) error {
	return service.connection.CreateObject(BucketName, func(id uint64) (int, interface{}) {
		cloudcredential.ID = models.CloudCredentialID(id)
		cloudcredential.Created = time.Now().Unix()
		return int(cloudcredential.ID), cloudcredential
	})
}

// GetNextIdentifier returns the next identifier for an cloudcredential.
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
