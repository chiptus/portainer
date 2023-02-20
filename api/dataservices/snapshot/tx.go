package snapshot

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
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

	err := service.tx.GetAllWithJsoniter(
		BucketName,
		&portaineree.Snapshot{},
		func(obj interface{}) (interface{}, error) {
			snapshot, ok := obj.(*portaineree.Snapshot)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to Snapshot object")
				return nil, fmt.Errorf("failed to convert to Snapshot object: %s", obj)
			}
			snapshots = append(snapshots, *snapshot)
			return &portaineree.Snapshot{}, nil
		})

	return snapshots, err
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
