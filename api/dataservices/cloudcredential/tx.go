package cloudcredential

import (
	"fmt"
	"time"

	"github.com/portainer/portainer-ee/api/database/models"
	portainer "github.com/portainer/portainer/api"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

// GetByID returns a cloudcredential by ID.
func (service ServiceTx) GetByID(ID models.CloudCredentialID) (*models.CloudCredential, error) {
	var cloudcredential *models.CloudCredential
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &cloudcredential)

	return cloudcredential, err
}

// Update updates a cloudcredential.
func (service ServiceTx) Update(ID models.CloudCredentialID, cloudcredential *models.CloudCredential) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, cloudcredential)
}

// Delete deletes a cloudcredential.
func (service ServiceTx) Delete(ID models.CloudCredentialID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}

// GetAll returns an array containing all the cloudcredentials.
func (service ServiceTx) GetAll() ([]models.CloudCredential, error) {
	var cloudcreds = make([]models.CloudCredential, 0)

	err := service.tx.GetAllWithJsoniter(
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

// Create assigns an ID to a new cloudcredential and saves it.
func (service ServiceTx) Create(cloudcredential *models.CloudCredential) error {
	return service.tx.CreateObject(BucketName, func(id uint64) (int, interface{}) {
		cloudcredential.ID = models.CloudCredentialID(id)
		cloudcredential.Created = time.Now().Unix()

		return int(cloudcredential.ID), cloudcredential
	})
}

// GetNextIdentifier returns the next identifier for a cloudcredential.
func (service ServiceTx) GetNextIdentifier() int {
	return service.tx.GetNextIdentifier(BucketName)
}
