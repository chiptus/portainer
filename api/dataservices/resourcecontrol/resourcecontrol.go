package resourcecontrol

import (
	"errors"
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

// BucketName represents the name of the bucket where this service stores data.
const BucketName = "resource_control"

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	dataservices.BaseDataService[portaineree.ResourceControl, portaineree.ResourceControlID]
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		BaseDataService: dataservices.BaseDataService[portaineree.ResourceControl, portaineree.ResourceControlID]{
			Bucket:     BucketName,
			Connection: connection,
		},
	}, nil
}

func (service *Service) Tx(tx portainer.Transaction) ServiceTx {
	return ServiceTx{
		BaseDataServiceTx: dataservices.BaseDataServiceTx[portaineree.ResourceControl, portaineree.ResourceControlID]{
			Bucket:     BucketName,
			Connection: service.Connection,
			Tx:         tx,
		},
	}
}

// ResourceControlByResourceIDAndType returns a ResourceControl object by checking if the resourceID is equal
// to the main ResourceID or in SubResourceIDs. It also performs a check on the resource type. Return nil
// if no ResourceControl was found.
func (service *Service) ResourceControlByResourceIDAndType(resourceID string, resourceType portaineree.ResourceControlType) (*portaineree.ResourceControl, error) {
	var resourceControl *portaineree.ResourceControl
	stop := fmt.Errorf("ok")
	err := service.Connection.GetAll(
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
func (service *Service) Create(resourceControl *portaineree.ResourceControl) error {
	return service.Connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			resourceControl.ID = portaineree.ResourceControlID(id)
			return int(resourceControl.ID), resourceControl
		},
	)
}
