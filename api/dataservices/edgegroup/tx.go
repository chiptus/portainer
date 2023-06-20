package edgegroup

import (
	"errors"

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

// EdgeGroups return a slice containing all the Edge groups.
func (service ServiceTx) EdgeGroups() ([]portaineree.EdgeGroup, error) {
	var groups = make([]portaineree.EdgeGroup, 0)

	return groups, service.tx.GetAllWithJsoniter(
		BucketName,
		&portaineree.EdgeGroup{},
		dataservices.AppendFn(&groups),
	)
}

// EdgeGroup returns an Edge group by ID.
func (service ServiceTx) EdgeGroup(ID portaineree.EdgeGroupID) (*portaineree.EdgeGroup, error) {
	var group portaineree.EdgeGroup
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

// UpdateEdgeGroup updates an edge group.
func (service ServiceTx) UpdateEdgeGroup(ID portaineree.EdgeGroupID, group *portaineree.EdgeGroup) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, group)
}

// UpdateEdgeGroupFunc is a no-op inside a transaction.
func (service ServiceTx) UpdateEdgeGroupFunc(ID portaineree.EdgeGroupID, updateFunc func(edgeGroup *portaineree.EdgeGroup)) error {
	return errors.New("cannot be called inside a transaction")
}

// DeleteEdgeGroup deletes an Edge group.
func (service ServiceTx) DeleteEdgeGroup(ID portaineree.EdgeGroupID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}

func (service ServiceTx) Create(group *portaineree.EdgeGroup) error {
	return service.tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			group.ID = portaineree.EdgeGroupID(id)
			return int(group.ID), group
		},
	)
}
