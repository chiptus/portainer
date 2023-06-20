package snapshot

import (
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

type ServiceTx struct {
	service *Service
	tx      portainer.Transaction
}

func (service ServiceTx) BucketName() string {
	return BucketName
}

func (service ServiceTx) Snapshot(endpointID portaineree.EndpointID) (*portaineree.Snapshot, error) {
	var snapshot portaineree.Snapshot
	identifier := service.service.connection.ConvertToKey(int(endpointID))

	err := service.tx.GetObject(BucketName, identifier, &snapshot)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (service ServiceTx) Snapshots() ([]portaineree.Snapshot, error) {
	var snapshots = make([]portaineree.Snapshot, 0)

	return snapshots, service.tx.GetAllWithJsoniter(
		BucketName,
		&portaineree.Snapshot{},
		dataservices.AppendFn(&snapshots),
	)
}

func (service ServiceTx) UpdateSnapshot(snapshot *portaineree.Snapshot) error {
	identifier := service.service.connection.ConvertToKey(int(snapshot.EndpointID))
	return service.tx.UpdateObject(BucketName, identifier, snapshot)
}

func (service ServiceTx) DeleteSnapshot(endpointID portaineree.EndpointID) error {
	identifier := service.service.connection.ConvertToKey(int(endpointID))
	return service.tx.DeleteObject(BucketName, identifier)
}

func (service ServiceTx) Create(snapshot *portaineree.Snapshot) error {
	return service.tx.CreateObjectWithId(BucketName, int(snapshot.EndpointID), snapshot)
}
