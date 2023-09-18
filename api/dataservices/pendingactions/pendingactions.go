package pendingactions

import (
	"time"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

const (
	BucketName = "pending_actions"
)

type Service struct {
	dataservices.BaseDataService[portaineree.PendingActions, portaineree.PendingActionsID]
}

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.PendingActions, portaineree.PendingActionsID]
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.PendingActions, portaineree.PendingActionsID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (s Service) Create(config *portaineree.PendingActions) error {
	return s.Connection.UpdateTx(func(tx portainer.Transaction) error {
		return s.Tx(tx).Create(config)
	})
}

func (s Service) Update(ID portaineree.PendingActionsID, config *portaineree.PendingActions) error {
	return s.Connection.UpdateTx(func(tx portainer.Transaction) error {
		return s.Tx(tx).Update(ID, config)
	})
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.PendingActions, portaineree.PendingActionsID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

func (s ServiceTx) Create(config *portaineree.PendingActions) error {
	return s.Tx.CreateObject(BucketName, func(id uint64) (int, interface{}) {
		config.ID = portaineree.PendingActionsID(id)
		config.CreatedAt = time.Now().Unix()

		return int(config.ID), config
	})
}

func (s ServiceTx) Update(ID portaineree.PendingActionsID, config *portaineree.PendingActions) error {
	return s.BaseDataServiceTx.Update(ID, config)
}

// GetNextIdentifier returns the next identifier for a custom template.
func (service *Service) GetNextIdentifier() int {
	return service.Connection.GetNextIdentifier(BucketName)
}
