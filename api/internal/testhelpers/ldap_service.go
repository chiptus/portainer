package testhelpers

import portaineree "github.com/portainer/portainer-ee/api"

type ldapService struct{}

// NewGitService creates new mock for portaineree.GitService.
func NewLDAPService() *ldapService {
	return &ldapService{}
}

// AuthenticateUser is used to authenticate a user against a LDAP/AD.
func (service *ldapService) AuthenticateUser(username, password string, settings *portaineree.LDAPSettings) error {
	return nil
}

// GetUserGroups is used to retrieve user groups from LDAP/AD.
func (service *ldapService) GetUserGroups(username string, settings *portaineree.LDAPSettings, useAutoAdminSearchSettings bool) ([]string, error) {
	if useAutoAdminSearchSettings {
		return []string{"stuff", "operator"}, nil
	}
	return []string{"stuff"}, nil
}

// SearchUsers searches for users with the specified settings
func (service *ldapService) SearchUsers(settings *portaineree.LDAPSettings) ([]string, error) {
	return nil, nil
}

// SearchGroups searches for groups with the specified settings
func (service *ldapService) SearchGroups(settings *portaineree.LDAPSettings) ([]portaineree.LDAPUser, error) {
	return nil, nil
}

// SearchGroups searches for groups with the specified settings
func (service *ldapService) SearchAdminGroups(settings *portaineree.LDAPSettings) ([]string, error) {
	return nil, nil
}

func (service *ldapService) TestConnectivity(settings *portaineree.LDAPSettings) error {
	return nil
}
