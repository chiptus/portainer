package resourcecontrol

import (
	"errors"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"

	"github.com/rs/zerolog/log"
)

type ServiceTx struct {
	dataservices.BaseDataServiceTx[portaineree.ResourceControl, portaineree.ResourceControlID]
}

// ResourceControlByResourceIDAndType returns a ResourceControl object by checking if the resourceID is equal
// to the main ResourceID or in SubResourceIDs. It also performs a check on the resource type. Return nil
// if no ResourceControl was found.
func (service ServiceTx) ResourceControlByResourceIDAndType(resourceID string, resourceType portaineree.ResourceControlType) (*portaineree.ResourceControl, error) {
	var resourceControl *portaineree.ResourceControl
	stop := fmt.Errorf("ok")
	err := service.Tx.GetAll(
		BucketName,
		&portaineree.ResourceControl{},
		func(obj interface{}) (interface{}, error) {
			rc, ok := obj.(*portaineree.ResourceControl)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to ResourceControl object")
				return nil, fmt.Errorf("failed to convert to ResourceControl object: %s", obj)
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

// CreateResourceControl creates a new ResourceControl object
func (service ServiceTx) Create(resourceControl *portaineree.ResourceControl) error {
	return service.Tx.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			resourceControl.ID = portaineree.ResourceControlID(id)
			return int(resourceControl.ID), resourceControl
		},
	)
}
