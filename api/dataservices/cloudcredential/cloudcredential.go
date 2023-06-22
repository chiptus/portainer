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
	dataservices.BaseDataService[models.CloudCredential, models.CloudCredentialID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[models.CloudCredential, models.CloudCredentialID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[models.CloudCredential, models.CloudCredentialID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// Create assigns an ID to a new cloudcredential and saves it.
func (service *Service) Create(cloudcredential *models.CloudCredential) error {
	return service.Connection.CreateObject(BucketName, func(id uint64) (int, interface{}) {
		cloudcredential.ID = models.CloudCredentialID(id)
		cloudcredential.Created = time.Now().Unix()
		return int(cloudcredential.ID), cloudcredential
	})
}

// GetNextIdentifier returns the next identifier for a cloudcredential.
func (service *Service) GetNextIdentifier() int {
	return service.Connection.GetNextIdentifier(BucketName)
}
