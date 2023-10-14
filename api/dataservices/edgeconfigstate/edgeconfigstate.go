package edgeconfigstate

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

const BucketName = "edgeconfigstates"

type Service struct {
	dataservices.BaseDataService[portaineree.EdgeConfigState, portainer.EndpointID]
}

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.EdgeConfigState, portainer.EndpointID]
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.EdgeConfigState, portainer.EndpointID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (s *Service) Create(state *portaineree.EdgeConfigState) error {
	return s.Connection.UpdateTx(func(tx portainer.Transaction) error {
		return s.Tx(tx).Create(state)
	})
}

func (s *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.EdgeConfigState, portainer.EndpointID]{
			Bucket:     BucketName,
			Connection: s.Connection,
			Tx:         tx,
		},
	}
}

func (s ServiceTx) Create(state *portaineree.EdgeConfigState) error {
	return s.Tx.CreateObjectWithId(BucketName, int(state.EndpointID), state)
}
