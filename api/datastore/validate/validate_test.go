package validate

import (
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
	"github.com/portainer/portainer-ee/api/internal/authorization"
)

func TestValidateLDAPSettings(t *testing.T) {

	tests := []struct {
		name    string
		ldap    portaineree.LDAPSettings
		wantErr bool
	}{
		{
			name:    "Empty LDAP Settings",
			ldap:    portaineree.LDAPSettings{},
			wantErr: true,
		},
		{
			name: "With URL but empty URLs",
			ldap: portaineree.LDAPSettings{
				AnonymousMode: true,
				URL:           "192.168.0.1:323",
			},
			wantErr: true,
		},
		{
			name: "Validate URL and URLs",
			ldap: portaineree.LDAPSettings{
				AnonymousMode: true,
				URL:           "192.168.0.1:323",
				URLs:          []string{":325"},
			},
			wantErr: false,
		},
		{
			name: "validate client ldap",
			ldap: portaineree.LDAPSettings{
				AnonymousMode: false,
				ReaderDN:      "CN=LDAP API Service Account",
				Password:      "Qu**dfUUU**",
				URLs: []string{
					"aukdc15.pgc.co:389",
				},
				TLSConfig: portaineree.TLSConfiguration{
					TLS:           false,
					TLSSkipVerify: false,
				},
				URL: "aukdc15.pgc.co:389",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLDAPSettings(&tt.ldap)
			if (err == nil) == tt.wantErr {
				t.Errorf("No error expected but got %s", err)
			}
		})
	}
}

func TestValidatePredefinedRoles(t *testing.T) {
	tests := []struct {
		name   string
		roles  []portaineree.Role
		errMsg string
	}{
		{
			name: "All roles",
			roles: []portaineree.Role{
				{
					Name:           "Environment administrator",
					Description:    "Full control of all resources in an environment",
					ID:             portaineree.RoleIDEndpointAdmin,
					Priority:       1,
					Authorizations: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
				},
				{
					Name:           "Helpdesk",
					Description:    "Read-only access of all resources in an environment",
					ID:             portaineree.RoleIDHelpdesk,
					Priority:       3,
					Authorizations: authorization.DefaultEndpointAuthorizationsForHelpDeskRole(),
				},
				{
					Name:           "Standard user",
					Description:    "Full control of assigned resources in an environment",
					ID:             portaineree.RoleIDStandardUser,
					Priority:       4,
					Authorizations: authorization.DefaultEndpointAuthorizationsForStandardUserRole(),
				},
			},
			errMsg: "predefined roles missing. Want=5, Got=3",
		},
		{
			name: "All roles",
			roles: []portaineree.Role{
				{
					Name:           "Environment administrator",
					Description:    "Full control of all resources in an environment",
					ID:             portaineree.RoleIDEndpointAdmin,
					Priority:       1,
					Authorizations: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
				},
				{
					Name:           "Helpdesk",
					Description:    "Read-only access of all resources in an environment",
					ID:             portaineree.RoleIDHelpdesk,
					Priority:       3,
					Authorizations: authorization.DefaultEndpointAuthorizationsForHelpDeskRole(),
				},
				{
					Name:           "Standard user",
					Description:    "Full control of assigned resources in an environment",
					ID:             portaineree.RoleIDStandardUser,
					Priority:       4,
					Authorizations: authorization.DefaultEndpointAuthorizationsForStandardUserRole(),
				},
				{
					Name:           "Read-only user",
					Description:    "Read-only access of assigned resources in an environment",
					ID:             portaineree.RoleIDReadonly,
					Priority:       5,
					Authorizations: authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole(),
				},
				{
					Name:           "Operator",
					Description:    "Operational control of all existing resources in an environment",
					ID:             portaineree.RoleIDOperator,
					Priority:       2,
					Authorizations: authorization.DefaultEndpointAuthorizationsForOperatorRole(),
				},
			},
			errMsg: "",
		},
		{
			name: "Operator role missing",
			roles: []portaineree.Role{
				{
					Name:           "Environment administrator",
					Description:    "Full control of all resources in an environment",
					ID:             portaineree.RoleIDEndpointAdmin,
					Priority:       1,
					Authorizations: authorization.DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
				},
				{
					Name:           "Helpdesk",
					Description:    "Read-only access of all resources in an environment",
					ID:             portaineree.RoleIDHelpdesk,
					Priority:       3,
					Authorizations: authorization.DefaultEndpointAuthorizationsForHelpDeskRole(),
				},
				{
					Name:           "Standard user",
					Description:    "Full control of assigned resources in an environment",
					ID:             portaineree.RoleIDStandardUser,
					Priority:       4,
					Authorizations: authorization.DefaultEndpointAuthorizationsForStandardUserRole(),
				},
				{
					Name:           "Read-only user",
					Description:    "Read-only access of assigned resources in an environment",
					ID:             portaineree.RoleIDReadonly,
					Priority:       5,
					Authorizations: authorization.DefaultEndpointAuthorizationsForReadOnlyUserRole(),
				},
				{
					Name:           "Operator 1",
					Description:    "Operational control of all existing resources in an environment",
					ID:             6,
					Priority:       2,
					Authorizations: authorization.DefaultEndpointAuthorizationsForOperatorRole(),
				},
			},
			errMsg: "role Operator missing in the DB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePredefinedRoles(tt.roles)
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("Want=%s, Got=%s", tt.errMsg, err)
			}
		})
	}
}
