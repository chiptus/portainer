package edgeupdateschedule

import (
	"github.com/portainer/portainer-ee/api/dataservices"
	edgetypes "github.com/portainer/portainer-ee/api/internal/edge/types"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[edgetypes.UpdateSchedule, edgetypes.UpdateScheduleID]
}

func (service ServiceTx) Create(item *edgetypes.UpdateSchedule) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			item.ID = edgetypes.UpdateScheduleID(id)
			return int(item.ID), item
		},
	)
}
