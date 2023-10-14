package snapshot

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

const BucketName = "snapshots"

type Service struct {
	dataservices.BaseDataService[portaineree.Snapshot, portainer.EndpointID]
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.Snapshot, portainer.EndpointID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.Snapshot, portainer.EndpointID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

func (service *Service) Create(snapshot *portaineree.Snapshot) error {
	return service.Connection.CreateObjectWithId(BucketName, int(snapshot.EndpointID), snapshot)
}
