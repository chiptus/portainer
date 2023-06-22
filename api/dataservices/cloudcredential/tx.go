package cloudcredential

import (
	"time"

	"github.com/portainer/portainer-ee/api/database/models"
	"github.com/portainer/portainer-ee/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[models.CloudCredential, models.CloudCredentialID]
}

// Create assigns an ID to a new cloudcredential and saves it.
func (service ServiceTx) Create(cloudcredential *models.CloudCredential) error {
	return service.Tx.CreateObject(BucketName, func(id uint64) (int, interface{}) {
		cloudcredential.ID = models.CloudCredentialID(id)
		cloudcredential.Created = time.Now().Unix()

		return int(cloudcredential.ID), cloudcredential
	})
}

// GetNextIdentifier returns the next identifier for a cloudcredential.
func (service ServiceTx) GetNextIdentifier() int {
	return service.Tx.GetNextIdentifier(BucketName)
}
