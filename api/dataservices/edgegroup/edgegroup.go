package edgegroup

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "edgegroups"

// Service represents a service for managing Edge group data.
type Service struct {
	dataservices.BaseDataService[portaineree.EdgeGroup, portaineree.EdgeGroupID]
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
		BaseDataService: dataservices.BaseDataService[portaineree.EdgeGroup, portaineree.EdgeGroupID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.EdgeGroup, portaineree.EdgeGroupID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// UpdateEdgeGroupFunc updates an edge group inside a transaction avoiding data races.
func (service *Service) UpdateEdgeGroupFunc(ID portaineree.EdgeGroupID, updateFunc func(edgeGroup *portaineree.EdgeGroup)) error {
	id := service.Connection.ConvertToKey(int(ID))
	edgeGroup := &portaineree.EdgeGroup{}

	return service.Connection.UpdateObjectFunc(BucketName, id, edgeGroup, func() {
		updateFunc(edgeGroup)
	})
}

// CreateEdgeGroup assign an ID to a new Edge group and saves it.
func (service *Service) Create(group *portaineree.EdgeGroup) error {
	return service.Connection.UpdateTx(func(tx portainer.Transaction) error {
		return service.Tx(tx).Create(group)
	})
}
