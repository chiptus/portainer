package edgejob

import (
	"errors"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

// EdgeJobs returns a list of Edge jobs
func (service ServiceTx) EdgeJobs() ([]portaineree.EdgeJob, error) {
	var edgeJobs = make([]portaineree.EdgeJob, 0)

	err := service.tx.GetAll(
		BucketName,
		&portaineree.EdgeJob{},
		func(obj interface{}) (interface{}, error) {
			job, ok := obj.(*portaineree.EdgeJob)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to EdgeJob object")
				return nil, fmt.Errorf("Failed to convert to EdgeJob object: %s", obj)
			}

			edgeJobs = append(edgeJobs, *job)

			return &portaineree.EdgeJob{}, nil
		})

	return edgeJobs, err
}

// EdgeJob returns an Edge job by ID
func (service ServiceTx) EdgeJob(ID portaineree.EdgeJobID) (*portaineree.EdgeJob, error) {
	var edgeJob portaineree.EdgeJob
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &edgeJob)
	if err != nil {
		return nil, err
	}

	return &edgeJob, nil
}

// Create creates a new EdgeJob
func (service ServiceTx) Create(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	edgeJob.ID = ID

	return service.tx.CreateObjectWithId(
		BucketName,
		int(edgeJob.ID),
		edgeJob,
	)
}

// UpdateEdgeJob updates an edge job
func (service ServiceTx) UpdateEdgeJob(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, edgeJob)
}

// UpdateEdgeJobFunc is a no-op inside a transaction
func (service ServiceTx) UpdateEdgeJobFunc(ID portaineree.EdgeJobID, updateFunc func(edgeJob *portaineree.EdgeJob)) error {
	return errors.New("cannot be called inside a transaction")
}

// DeleteEdgeJob deletes an Edge job
func (service ServiceTx) DeleteEdgeJob(ID portaineree.EdgeJobID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service ServiceTx) GetNextIdentifier() int {
	return service.tx.GetNextIdentifier(BucketName)
}
