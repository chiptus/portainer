package migrator

import (
	"github.com/boltdb/bolt"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/internal"
)

func (m *Migrator) updateResourceControlsToDBVersion2() error {
	legacyResourceControls, err := m.retrieveLegacyResourceControls()
	if err != nil {
		return err
	}

	for _, resourceControl := range legacyResourceControls {
		resourceControl.SubResourceIDs = []string{}
		resourceControl.TeamAccesses = []portaineree.TeamResourceAccess{}

		owner, err := m.userService.User(resourceControl.OwnerID)
		if err != nil {
			return err
		}

		if owner.Role == portaineree.AdministratorRole {
			resourceControl.AdministratorsOnly = true
			resourceControl.UserAccesses = []portaineree.UserResourceAccess{}
		} else {
			resourceControl.AdministratorsOnly = false
			userAccess := portaineree.UserResourceAccess{
				UserID:      resourceControl.OwnerID,
				AccessLevel: portaineree.ReadWriteAccessLevel,
			}
			resourceControl.UserAccesses = []portaineree.UserResourceAccess{userAccess}
		}

		err = m.resourceControlService.CreateResourceControl(&resourceControl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) updateEndpointsToDBVersion2() error {
	legacyEndpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range legacyEndpoints {
		endpoint.AuthorizedTeams = []portaineree.TeamID{}
		err = m.endpointService.UpdateEndpoint(endpoint.ID, &endpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) retrieveLegacyResourceControls() ([]portaineree.ResourceControl, error) {
	legacyResourceControls := make([]portaineree.ResourceControl, 0)
	err := m.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("containerResourceControl"))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var resourceControl portaineree.ResourceControl
			err := internal.UnmarshalObject(v, &resourceControl)
			if err != nil {
				return err
			}
			resourceControl.Type = portaineree.ContainerResourceControl
			legacyResourceControls = append(legacyResourceControls, resourceControl)
		}

		bucket = tx.Bucket([]byte("serviceResourceControl"))
		cursor = bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var resourceControl portaineree.ResourceControl
			err := internal.UnmarshalObject(v, &resourceControl)
			if err != nil {
				return err
			}
			resourceControl.Type = portaineree.ServiceResourceControl
			legacyResourceControls = append(legacyResourceControls, resourceControl)
		}

		bucket = tx.Bucket([]byte("volumeResourceControl"))
		cursor = bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var resourceControl portaineree.ResourceControl
			err := internal.UnmarshalObject(v, &resourceControl)
			if err != nil {
				return err
			}
			resourceControl.Type = portaineree.VolumeResourceControl
			legacyResourceControls = append(legacyResourceControls, resourceControl)
		}
		return nil
	})
	return legacyResourceControls, err
}
