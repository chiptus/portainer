package edgejob

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "edgejobs"

// Service represents a service for managing edge jobs data.
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

// EdgeJobs returns a list of Edge jobs
func (service *Service) EdgeJobs() ([]portaineree.EdgeJob, error) {
	var edgeJobs = make([]portaineree.EdgeJob, 0)

	return edgeJobs, service.connection.GetAll(
		BucketName,
		&portaineree.EdgeJob{},
		dataservices.AppendFn(&edgeJobs),
	)
}

// EdgeJob returns an Edge job by ID
func (service *Service) EdgeJob(ID portaineree.EdgeJobID) (*portaineree.EdgeJob, error) {
	var edgeJob portaineree.EdgeJob
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &edgeJob)
	if err != nil {
		return nil, err
	}

	return &edgeJob, nil
}

// Create creates a new EdgeJob
func (service *Service) Create(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	edgeJob.ID = ID

	return service.connection.CreateObjectWithId(
		BucketName,
		int(edgeJob.ID),
		edgeJob,
	)
}

// Deprecated: Use UpdateEdgeJobFunc instead
func (service *Service) UpdateEdgeJob(ID portaineree.EdgeJobID, edgeJob *portaineree.EdgeJob) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, edgeJob)
}

// UpdateEdgeJobFunc updates an edge job inside a transaction avoiding data races.
func (service *Service) UpdateEdgeJobFunc(ID portaineree.EdgeJobID, updateFunc func(edgeJob *portaineree.EdgeJob)) error {
	id := service.connection.ConvertToKey(int(ID))
	edgeJob := &portaineree.EdgeJob{}

	return service.connection.UpdateObjectFunc(BucketName, id, edgeJob, func() {
		updateFunc(edgeJob)
	})
}

// DeleteEdgeJob deletes an Edge job
func (service *Service) DeleteEdgeJob(ID portaineree.EdgeJobID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}

// GetNextIdentifier returns the next identifier for an environment(endpoint).
func (service *Service) GetNextIdentifier() int {
	return service.connection.GetNextIdentifier(BucketName)
}
