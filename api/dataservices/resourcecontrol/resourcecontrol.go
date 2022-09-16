package resourcecontrol

import (
	"fmt"

	portaineree "github.com/portainer/portainer-ee/api"
	portainer "github.com/portainer/portainer/api"

	"github.com/rs/zerolog/log"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "resource_control"
)

// Service represents a service for managing environment(endpoint) data.
type Service struct {
	connection portainer.Connection
}

func (service *Service) BucketName() string {
	return BucketName
}

// NewService creates a new instance of a service.
func NewService(connection portainer.Connection) (*Service, error) {
	err := connection.SetServiceName(BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// ResourceControl returns a ResourceControl object by ID
func (service *Service) ResourceControl(ID portaineree.ResourceControlID) (*portaineree.ResourceControl, error) {
	var resourceControl portaineree.ResourceControl
	identifier := service.connection.ConvertToKey(int(ID))

	err := service.connection.GetObject(BucketName, identifier, &resourceControl)
	if err != nil {
		return nil, err
	}

	return &resourceControl, nil
}

// ResourceControlByResourceIDAndType returns a ResourceControl object by checking if the resourceID is equal
// to the main ResourceID or in SubResourceIDs. It also performs a check on the resource type. Return nil
// if no ResourceControl was found.
func (service *Service) ResourceControlByResourceIDAndType(resourceID string, resourceType portaineree.ResourceControlType) (*portaineree.ResourceControl, error) {
	var resourceControl *portaineree.ResourceControl
	stop := fmt.Errorf("ok")
	err := service.connection.GetAll(
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
	if err == stop {
		return resourceControl, nil
	}

	return nil, err
}

// ResourceControls returns all the ResourceControl objects
func (service *Service) ResourceControls() ([]portaineree.ResourceControl, error) {
	var rcs = make([]portaineree.ResourceControl, 0)

	err := service.connection.GetAll(
		BucketName,
		&portaineree.ResourceControl{},
		func(obj interface{}) (interface{}, error) {
			rc, ok := obj.(*portaineree.ResourceControl)
			if !ok {
				log.Debug().Str("obj", fmt.Sprintf("%#v", obj)).Msg("failed to convert to ResourceControl object")
				return nil, fmt.Errorf("Failed to convert to ResourceControl object: %s", obj)
			}

			rcs = append(rcs, *rc)

			return &portaineree.ResourceControl{}, nil
		})

	return rcs, err
}

// CreateResourceControl creates a new ResourceControl object
func (service *Service) Create(resourceControl *portaineree.ResourceControl) error {
	return service.connection.CreateObject(
		BucketName,
		func(id uint64) (int, interface{}) {
			resourceControl.ID = portaineree.ResourceControlID(id)
			return int(resourceControl.ID), resourceControl
		},
	)
}

// UpdateResourceControl saves a ResourceControl object.
func (service *Service) UpdateResourceControl(ID portaineree.ResourceControlID, resourceControl *portaineree.ResourceControl) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.UpdateObject(BucketName, identifier, resourceControl)
}

// DeleteResourceControl deletes a ResourceControl object by ID
func (service *Service) DeleteResourceControl(ID portaineree.ResourceControlID) error {
	identifier := service.connection.ConvertToKey(int(ID))
	return service.connection.DeleteObject(BucketName, identifier)
}
