package resourcecontrol

import (
	"errors"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
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

// ResourceControl returns a ResourceControl object by ID
func (service ServiceTx) ResourceControl(ID portaineree.ResourceControlID) (*portaineree.ResourceControl, error) {
	var resourceControl portaineree.ResourceControl
	identifier := service.service.connection.ConvertToKey(int(ID))

	err := service.tx.GetObject(BucketName, identifier, &resourceControl)
	if err != nil {
		return nil, err
	}

	return &resourceControl, nil
}

// ResourceControlByResourceIDAndType returns a ResourceControl object by checking if the resourceID is equal
// to the main ResourceID or in SubResourceIDs. It also performs a check on the resource type. Return nil
// if no ResourceControl was found.
func (service ServiceTx) ResourceControlByResourceIDAndType(resourceID string, resourceType portaineree.ResourceControlType) (*portaineree.ResourceControl, error) {
	var resourceControl *portaineree.ResourceControl
	stop := fmt.Errorf("ok")
	err := service.tx.GetAll(
		BucketName,
		&portaineree.ResourceControl{},
		func(obj interface{}) (interface{}, error) {
			rc, ok := obj.(*portaineree.ResourceControl)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to ResourceControl object")
				return nil, fmt.Errorf("Failed to convert to ResourceControl object: %s", obj)
			}

			if rc.ResourceID == resourceID && rc.Type == resourceType {
				resourceControl = rc
				return nil, stop
			}

			for _, subResourceID := range rc.SubResourceIDs {
				if subResourceID == resourceID {
					resourceControl = rc
					return nil, stop
				}
			}

			return &portaineree.ResourceControl{}, nil
		})
	if errors.Is(err, stop) {
		return resourceControl, nil
	}

	return nil, err
}

// ResourceControls returns all the ResourceControl objects
func (service ServiceTx) ResourceControls() ([]portaineree.ResourceControl, error) {
	var rcs = make([]portaineree.ResourceControl, 0)

	return rcs, service.tx.GetAll(
		BucketName,
		&portaineree.ResourceControl{},
		dataservices.AppendFn(&rcs),
	)
}

// CreateResourceControl creates a new ResourceControl object
func (service ServiceTx) Create(resourceControl *portaineree.ResourceControl) error {
	return service.tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			resourceControl.ID = portaineree.ResourceControlID(id)
			return int(resourceControl.ID), resourceControl
		},
	)
}

// UpdateResourceControl saves a ResourceControl object.
func (service ServiceTx) UpdateResourceControl(ID portaineree.ResourceControlID, resourceControl *portaineree.ResourceControl) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.UpdateObject(BucketName, identifier, resourceControl)
}

// DeleteResourceControl deletes a ResourceControl object by ID
func (service ServiceTx) DeleteResourceControl(ID portaineree.ResourceControlID) error {
	identifier := service.service.connection.ConvertToKey(int(ID))
	return service.tx.DeleteObject(BucketName, identifier)
}
