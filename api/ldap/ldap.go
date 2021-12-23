package ldap

import (
	"fmt"
	"log"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
	"github.com/pkg/errors"
	portaineree "github.com/portainer/portainer-ee/api"
	httperrors "github.com/portainer/portainer-ee/api/http/errors"
	"github.com/portainer/portainer/api/crypto"
)

var (
	// errUserNotFound defines an error raised when the user is not found via LDAP search
	// or that too many entries (> 1) are returned.
	errUserNotFound = errors.New("User not found or too many entries returned")
)

// Service represents a service used to authenticate users against a LDAP/AD.
type Service struct{}

func createConnection(settings *portaineree.LDAPSettings) (*ldap.Conn, error) {
	for _, url := range settings.URLs {
		conn, err := createConnectionForURL(url, settings)
		if err != nil {
			log.Printf("[DEBUG] [ldap] [message: failed creating LDAP connection] [error: %s]", err)
		} else {
			return conn, nil
		}
	}

	return nil, errors.New("No valid connection")
}

func createConnectionForURL(url string, settings *portaineree.LDAPSettings) (*ldap.Conn, error) {
	if settings.TLSConfig.TLS || settings.StartTLS {
		config, err := crypto.CreateTLSConfigurationFromDisk(settings.TLSConfig.TLSCACertPath, settings.TLSConfig.TLSCertPath, settings.TLSConfig.TLSKeyPath, settings.TLSConfig.TLSSkipVerify)
		if err != nil {
			return nil, err
		}
		config.ServerName = strings.Split(url, ":")[0]

		if settings.TLSConfig.TLS {
			return ldap.DialTLS("tcp", url, config)
		}

		conn, err := ldap.Dial("tcp", url)
		if err != nil {
			return nil, err
		}

		err = conn.StartTLS(config)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	return ldap.Dial("tcp", url)
}

// AuthenticateUser is used to authenticate a user against a LDAP/AD.
func (*Service) AuthenticateUser(username, password string, settings *portaineree.LDAPSettings) error {

	connection, err := createConnection(settings)
	if err != nil {
		return err
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		err = connection.Bind(settings.ReaderDN, settings.Password)
		if err != nil {
			return err
		}
	}

	userDN, err := searchUser(username, connection, settings.SearchSettings)
	if err != nil {
		return err
	}

	err = connection.Bind(userDN, password)
	if err != nil {
		return httperrors.ErrUnauthorized
	}

	return nil
}

// GetUserGroups is used to retrieve user groups from LDAP/AD.
func (*Service) GetUserGroups(username string, settings *portaineree.LDAPSettings, useAutoAdminSearchSettings bool) ([]string, error) {
	connection, err := createConnection(settings)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		err = connection.Bind(settings.ReaderDN, settings.Password)
		if err != nil {
			return nil, err
		}
	}

	userDN, err := searchUser(username, connection, settings.SearchSettings)
	if err != nil {
		return nil, err
	}

	groupSearchSettings := settings.GroupSearchSettings
	if useAutoAdminSearchSettings {
		groupSearchSettings = settings.AdminGroupSearchSettings
	}

	userGroups := getGroupsByUser(userDN, connection, groupSearchSettings)

	return userGroups, nil
}

// SearchUsers searches for users with the specified settings
func (*Service) SearchUsers(settings *portaineree.LDAPSettings) ([]string, error) {
	connection, err := createConnection(settings)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		err = connection.Bind(settings.ReaderDN, settings.Password)
		if err != nil {
			return nil, err
		}
	}

	users := map[string]bool{}

	for _, searchSettings := range settings.SearchSettings {
		searchRequest := ldap.NewSearchRequest(
			searchSettings.BaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			searchSettings.Filter,
			[]string{"dn", searchSettings.UserNameAttribute},
			nil,
		)

		sr, err := connection.Search(searchRequest)
		if err != nil {
			return nil, err
		}

		for _, user := range sr.Entries {
			username := user.GetAttributeValue(searchSettings.UserNameAttribute)
			if username != "" {
				users[username] = true
			}
		}
	}

	usersList := []string{}
	for user := range users {
		usersList = append(usersList, user)
	}

	return usersList, nil
}

// SearchGroups searches for groups with the GroupSearchSettings
func (*Service) SearchGroups(settings *portaineree.LDAPSettings) ([]portaineree.LDAPUser, error) {
	userGroups, err := searchUserGroups(settings, false)
	if err != nil {
		return nil, errors.WithMessage(err, "failed searching user groups")
	}

	users := []portaineree.LDAPUser{}

	for username, groups := range userGroups {
		user := portaineree.LDAPUser{
			Name:   username,
			Groups: groups,
		}
		users = append(users, user)
	}

	return users, nil
}

// SearchAdminGroups searches for groups with the AdminGroupSearchSettings
func (*Service) SearchAdminGroups(settings *portaineree.LDAPSettings) ([]string, error) {
	userGroups, err := searchUserGroups(settings, true)
	if err != nil {
		return nil, errors.WithMessage(err, "failed searching user groups")
	}

	deduplicatedGroups := make(map[string]struct{})
	for _, gs := range userGroups {
		for _, group := range gs {
			deduplicatedGroups[group] = struct{}{}
		}
	}
	groups := make([]string, 0, len(deduplicatedGroups))
	for group := range deduplicatedGroups {
		groups = append(groups, group)
	}

	return groups, nil
}

func searchUser(username string, conn *ldap.Conn, settings []portaineree.LDAPSearchSettings) (string, error) {
	var userDN string
	found := false
	usernameEscaped := ldap.EscapeFilter(username)

	for _, searchSettings := range settings {
		searchRequest := ldap.NewSearchRequest(
			searchSettings.BaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%s(%s=%s))", searchSettings.Filter, searchSettings.UserNameAttribute, usernameEscaped),
			[]string{"dn"},
			nil,
		)

		// Deliberately skip errors on the search request so that we can jump to other search settings
		// if any issue arise with the current one.
		sr, err := conn.Search(searchRequest)
		if err != nil {
			continue
		}

		if len(sr.Entries) == 1 {
			found = true
			userDN = sr.Entries[0].DN
			break
		}
	}

	if !found {
		return "", errUserNotFound
	}

	return userDN, nil
}

func searchUserGroups(settings *portaineree.LDAPSettings, useAutoAdminSearchSettings bool) (map[string][]string, error) {
	connection, err := createConnection(settings)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to esteblish an LDAP connection")
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		if err := connection.Bind(settings.ReaderDN, settings.Password); err != nil {
			return nil, errors.Wrap(err, "failed to bind an LDAP connection")
		}
	}

	groupSearchSettings := settings.GroupSearchSettings
	if useAutoAdminSearchSettings {
		groupSearchSettings = settings.AdminGroupSearchSettings
	}

	userGroups := make(map[string][]string)

	for _, searchSettings := range groupSearchSettings {
		searchRequest := ldap.NewSearchRequest(
			searchSettings.GroupBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			searchSettings.GroupFilter,
			[]string{"cn", searchSettings.GroupAttribute},
			nil,
		)

		sr, err := connection.Search(searchRequest)
		if err != nil {
			return nil, errors.Wrap(err, "failed to perform a user groups search")
		}

		for _, entry := range sr.Entries {
			members := entry.GetAttributeValues(searchSettings.GroupAttribute)
			for _, username := range members {
				userGroups[username] = append(userGroups[username], entry.GetAttributeValue("cn"))
			}
		}
	}

	return userGroups, nil
}

// Get a list of group names for specified user from LDAP/AD
func getGroupsByUser(userDN string, conn *ldap.Conn, settings []portaineree.LDAPGroupSearchSettings) []string {
	groups := make([]string, 0)
	userDNEscaped := ldap.EscapeFilter(userDN)

	for _, searchSettings := range settings {
		searchRequest := ldap.NewSearchRequest(
			searchSettings.GroupBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%s(%s=%s))", searchSettings.GroupFilter, searchSettings.GroupAttribute, userDNEscaped),
			[]string{"cn"},
			nil,
		)

		// Deliberately skip errors on the search request so that we can jump to other search settings
		// if any issue arise with the current one.
		sr, err := conn.Search(searchRequest)
		if err != nil {
			continue
		}

		for _, entry := range sr.Entries {
			for _, attr := range entry.Attributes {
				groups = append(groups, attr.Values[0])
			}
		}
	}

	return groups
}

// TestConnectivity is used to test a connection against the LDAP server using the credentials
// specified in the LDAPSettings.
func (*Service) TestConnectivity(settings *portaineree.LDAPSettings) error {

	connection, err := createConnection(settings)
	if err != nil {
		return err
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		err = connection.Bind(settings.ReaderDN, settings.Password)
		if err != nil {
			return err
		}

	} else {
		err = connection.UnauthenticatedBind("")
		if err != nil {
			return err
		}
	}

	return nil
}
