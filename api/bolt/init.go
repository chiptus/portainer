package bolt

import (
	"github.com/gofrs/uuid"
	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/bolt/errors"
)

// Init creates the default data set.
func (store *Store) Init() error {
	instanceID, err := store.VersionService.InstanceID()
	if err == errors.ErrObjectNotFound {
		uid, err := uuid.NewV4()
		if err != nil {
			return err
		}

		instanceID = uid.String()
		err = store.VersionService.StoreInstanceID(instanceID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	_, err = store.SettingsService.Settings()
	if err == errors.ErrObjectNotFound {
		defaultSettings := &portaineree.Settings{
			AuthenticationMethod: portaineree.AuthenticationInternal,
			BlackListedLabels:    make([]portaineree.Pair, 0),
			LDAPSettings: portaineree.LDAPSettings{
				AnonymousMode:   true,
				AutoCreateUsers: true,
				TLSConfig:       portaineree.TLSConfiguration{},
				URLs:            []string{},
				SearchSettings: []portaineree.LDAPSearchSettings{
					{},
				},
				GroupSearchSettings: []portaineree.LDAPGroupSearchSettings{
					{},
				},
				AdminGroupSearchSettings: []portaineree.LDAPGroupSearchSettings{
					{},
				},
			},
			OAuthSettings: portaineree.OAuthSettings{
				TeamMemberships: portaineree.TeamMemberships{
					OAuthClaimMappings: make([]portaineree.OAuthClaimMappings, 0),
				},
			},
			EdgeAgentCheckinInterval: portaineree.DefaultEdgeAgentCheckinIntervalInSeconds,
			TemplatesURL:             portaineree.DefaultTemplatesURL,
			HelmRepositoryURL:        portaineree.DefaultHelmRepositoryURL,
			UserSessionTimeout:       portaineree.DefaultUserSessionTimeout,
			KubeconfigExpiry:         portaineree.DefaultKubeconfigExpiry,
			KubectlShellImage:        portaineree.DefaultKubectlShellImage,
		}

		err = store.SettingsService.UpdateSettings(defaultSettings)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	_, err = store.SSLSettings().Settings()
	if err != nil {
		if err != errors.ErrObjectNotFound {
			return err
		}

		defaultSSLSettings := &portaineree.SSLSettings{
			HTTPEnabled: true,
		}

		err = store.SSLSettings().UpdateSettings(defaultSSLSettings)
		if err != nil {
			return err
		}
	}

	groups, err := store.EndpointGroupService.EndpointGroups()
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		unassignedGroup := &portaineree.EndpointGroup{
			Name:               "Unassigned",
			Description:        "Unassigned environments",
			Labels:             []portaineree.Pair{},
			UserAccessPolicies: portaineree.UserAccessPolicies{},
			TeamAccessPolicies: portaineree.TeamAccessPolicies{},
			TagIDs:             []portaineree.TagID{},
		}

		err = store.EndpointGroupService.CreateEndpointGroup(unassignedGroup)
		if err != nil {
			return err
		}
	}

	roles, err := store.RoleService.Roles()
	if err != nil {
		return err
	}

	if len(roles) == 0 {
		err := store.RoleService.CreateOrUpdatePredefinedRoles()
		if err != nil {
			return err
		}
	}

	return nil
}
