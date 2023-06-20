package cloudcredential

import (
	"time"

	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "cloudcredentials"

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

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		service: service,
		tx:      tx,
	}
}

// GetByID returns a cloudcredential by ID.
func (service *Service) GetByID(ID models.CloudCredentialID) (*models.CloudCredential, error) {
	var cloudcredential models.CloudCredential
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &cloudcredential)
	if err != nil {
		return nil, err
	}

	return &cloudcredential, nil
}

// Update updates a cloudcredential.
func (service *Service) Update(ID models.CloudCredentialID, cloudcredential *models.CloudCredential) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, cloudcredential)
}

// Delete deletes a cloudcredential.
func (service *Service) Delete(ID models.CloudCredentialID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// GetAll returns an array containing all the cloudcredentials.
func (service *Service) GetAll() ([]models.CloudCredential, error) {
	var cloudcreds = make([]models.CloudCredential, 0)

	return cloudcreds, service.connection.GetAllWithJsoniter(
		BucketName,
		&models.CloudCredential{},
		dataservices.AppendFn(&cloudcreds),
	)
}

// Create assigns an ID to a new cloudcredential and saves it.
func (service *Service) Create(cloudcredential *models.CloudCredential) error {
	return service.connection.CreateObject(BucketName, func(id uint64) (int, interface{}) {
		cloudcredential.ID = models.CloudCredentialID(id)
		cloudcredential.Created = time.Now().Unix()
		return int(cloudcredential.ID), cloudcredential
	})
}

// GetNextIdentifier returns the next identifier for a cloudcredential.
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
