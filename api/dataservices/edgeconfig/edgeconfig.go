package edgeconfig

import (
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

const (
	BucketName = "edgeconfigs"

	EdgeConfigTypeGeneral portaineree.EdgeConfigType = iota
	EdgeConfigTypeSpecificFile
	EdgeConfigTypeSpecificFolder
)

type Service struct {
	dataservices.BaseDataService[portaineree.EdgeConfig, portaineree.EdgeConfigID]
}

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.EdgeConfig, portaineree.EdgeConfigID]
}

var EdgeConfigTypes = []portaineree.EdgeConfigType{
	EdgeConfigTypeGeneral,
	EdgeConfigTypeSpecificFile,
	EdgeConfigTypeSpecificFolder,
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.EdgeConfig, portaineree.EdgeConfigID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (s Service) Create(config *portaineree.EdgeConfig) error {
	return s.Connection.UpdateTx(func(tx portainer.Transaction) error {
		return s.Tx(tx).Create(config)
	})
}

func (s Service) Update(ID portaineree.EdgeConfigID, config *portaineree.EdgeConfig) error {
	return s.Connection.UpdateTx(func(tx portainer.Transaction) error {
		return s.Tx(tx).Update(ID, config)
	})
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.EdgeConfig, portaineree.EdgeConfigID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

func (s ServiceTx) Create(config *portaineree.EdgeConfig) error {
	return s.Tx.CreateObject(BucketName, func(id uint64) (int, interface{}) {
		config.ID = portaineree.EdgeConfigID(id)
		config.Created = time.Now().Unix()

		return int(config.ID), config
	})
}

func (s ServiceTx) Update(ID portaineree.EdgeConfigID, config *portaineree.EdgeConfig) error {
	config.Updated = time.Now().Unix()

	return s.BaseDataServiceTx.Update(ID, config)
}
