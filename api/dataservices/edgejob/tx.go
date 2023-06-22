package edgejob

import (
	"errors"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.EdgeJob, portaineree.EdgeJobID]
}

// Create creates a new EdgeJob
func (service ServiceTx) Create(edgeJob *portaineree.EdgeJob) error {
	return service.CreateWithID(portaineree.EdgeJobID(service.GetNextIdentifier()), edgeJob)
}

// CreateWithID creates a new EdgeJob
func (service ServiceTx) CreateWithID(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	edgeJob.ID = ID

	return service.Tx.CreateObjectWithId(
		BucketName,
		int(edgeJob.ID),
		edgeJob,
	)
}

// UpdateEdgeJobFunc is a no-op inside a transaction
func (service ServiceTx) UpdateEdgeJobFunc(ID portaineree.EdgeJobID, updateFunc func(edgeJob *portaineree.EdgeJob)) error {
	return errors.New("cannot be called inside a transaction")
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service ServiceTx) GetNextIdentifier() int {
	return service.Tx.GetNextIdentifier(BucketName)
}
