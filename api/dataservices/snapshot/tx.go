package snapshot

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.Snapshot, portainer.EndpointID]
}

func (service ServiceTx) Create(snapshot *portaineree.Snapshot) error {
	return service.Tx.CreateObjectWithId(BucketName, int(snapshot.EndpointID), snapshot)
}
