package endpointgroup

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

// EndpointGroup returns an environment(endpoint) group by ID.
func (service ServiceTx) EndpointGroup(ID portaineree.EndpointGroupID) (*portaineree.EndpointGroup, error) {
	var endpointGroup portaineree.EndpointGroup
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &endpointGroup)
	if err != nil {
		return nil, err
	}

	return &endpointGroup, nil
}

// UpdateEndpointGroup updates an environment(endpoint) group.
func (service ServiceTx) UpdateEndpointGroup(ID portaineree.EndpointGroupID, endpointGroup *portaineree.EndpointGroup) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, endpointGroup)
}

// DeleteEndpointGroup deletes an environment(endpoint) group.
func (service ServiceTx) DeleteEndpointGroup(ID portaineree.EndpointGroupID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}

// EndpointGroups return an array containing all the environment(endpoint) groups.
func (service ServiceTx) EndpointGroups() ([]portaineree.EndpointGroup, error) {
	var endpointGroups = make([]portaineree.EndpointGroup, 0)

	err := service.tx.GetAll(
		BucketName,
		&portaineree.EndpointGroup{},
		func(obj interface{}) (interface{}, error) {
			endpointGroup, ok := obj.(*portaineree.EndpointGroup)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to EndpointGroup object")
				return nil, fmt.Errorf("Failed to convert to EndpointGroup object: %s", obj)
			}

			endpointGroups = append(endpointGroups, *endpointGroup)

			return &portaineree.EndpointGroup{}, nil
		})

	return endpointGroups, err
}

// CreateEndpointGroup assign an ID to a new environment(endpoint) group and saves it.
func (service ServiceTx) Create(endpointGroup *portaineree.EndpointGroup) error {
	return service.tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			endpointGroup.ID = portaineree.EndpointGroupID(id)
			return int(endpointGroup.ID), endpointGroup
		},
	)
}
