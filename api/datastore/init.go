package datastore

import (
	portaineree "github.com/portainer/portainer-ee/api"
)

// Init creates the default data set.
func (store *Store) Init() error {
	err := store.checkOrCreateDefaultSettings()
	if err != nil {
		return err
	}

	err = store.checkOrCreateDefaultSSLSettings()
	if err != nil {
		return err
	}

	return store.checkOrCreateDefaultData()
}

func (store *Store) checkOrCreateDefaultSettings() error {
	_, err := store.SettingsService.Settings()
	if store.IsErrObjectNotFound(err) {
		defaultSettings := &portaineree.Settings{
			EnableTelemetry:      false,
			AuthenticationMethod: portaineree.AuthenticationInternal,
			InternalAuthSettings: portaineree.InternalAuthSettings{
				RequiredPasswordLength: 12,
			},
			BlackListedLabels: make([]portaineree.Pair, 0),
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
				SSO: true,
			},
			Edge: portaineree.Edge{
				CommandInterval:  60,
				PingInterval:     60,
				SnapshotInterval: 60,
			},
			ExperimentalFeatures: portaineree.ExperimentalFeatures{
				OpenAIIntegration: false,
			},
			SnapshotInterval:         portaineree.DefaultSnapshotInterval,
			EdgeAgentCheckinInterval: portaineree.DefaultEdgeAgentCheckinIntervalInSeconds,
			TemplatesURL:             portaineree.DefaultTemplatesURL,
			HelmRepositoryURL:        portaineree.DefaultHelmRepositoryURL,
			UserSessionTimeout:       portaineree.DefaultUserSessionTimeout,
			KubeconfigExpiry:         portaineree.DefaultKubeconfigExpiry,
			KubectlShellImage:        portaineree.DefaultKubectlShellImage,
		}

		return store.SettingsService.UpdateSettings(defaultSettings)
	}

	return err
}

func (store *Store) checkOrCreateDefaultSSLSettings() error {
	_, err := store.SSLSettings().Settings()
	if store.IsErrObjectNotFound(err) {
		defaultSSLSettings := &portaineree.SSLSettings{
			HTTPEnabled: true,
		}

		return store.SSLSettings().UpdateSettings(defaultSSLSettings)
	}

	return err
}

func (store *Store) checkOrCreateDefaultData() error {
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

		err = store.EndpointGroupService.Create(unassignedGroup)
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
