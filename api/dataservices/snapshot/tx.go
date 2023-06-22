package snapshot

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.Snapshot, portaineree.EndpointID]
}

func (service ServiceTx) Create(snapshot *portaineree.Snapshot) error {
	return service.Tx.CreateObjectWithId(BucketName, int(snapshot.EndpointID), snapshot)
}
