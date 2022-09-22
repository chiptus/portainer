package snapshot

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"
	"github.com/sirupsen/logrus"
)

const (
	BucketName = "snapshots"
)

type Service struct {
	connection portainer.Connection
}

func (service *Service) BucketName() string {
	return BucketName
}

func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

func (service *Service) Snapshot(endpointID portaineree.EndpointID) (*portaineree.Snapshot, error) {
	var snapshot portaineree.Snapshot
	identifier := service.connection.ConvertToKey(int(endpointID))

	err := service.connection.GetObject(BucketName, identifier, &snapshot)
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (service *Service) Snapshots() ([]portaineree.Snapshot, error) {
	var snapshots = make([]portaineree.Snapshot, 0)

	err := service.connection.GetAllWithJsoniter(
		BucketName,
		&portaineree.Snapshot{},
		func(obj interface{}) (interface{}, error) {
			snapshot, ok := obj.(*portaineree.Snapshot)
			if !ok {
				logrus.WithField("obj", obj).Errorf("Failed to convert to Snapshot object")
				return nil, fmt.Errorf("failed to convert to Snapshot object: %s", obj)
			}
			snapshots = append(snapshots, *snapshot)
			return &portaineree.Snapshot{}, nil
		})

	return snapshots, err
}

func (service *Service) UpdateSnapshot(snapshot *portaineree.Snapshot) error {
	identifier := service.connection.ConvertToKey(int(snapshot.EndpointID))
	return service.connection.UpdateObject(BucketName, identifier, snapshot)
}

func (service *Service) DeleteSnapshot(endpointID portaineree.EndpointID) error {
	identifier := service.connection.ConvertToKey(int(endpointID))
	return service.connection.DeleteObject(BucketName, identifier)
}

func (service *Service) Create(snapshot *portaineree.Snapshot) error {
	return service.connection.CreateObjectWithId(BucketName, int(snapshot.EndpointID), snapshot)
}
