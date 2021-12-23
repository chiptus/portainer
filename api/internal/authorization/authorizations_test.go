package authorization

import (
	"reflect"
	"testing"

	portaineree "github.com/portainer/portainer-ee/api"
)

func Test_getKeyRole(t *testing.T) {
	type args struct {
		roleIdentifiers []portaineree.RoleID
		roles           []portaineree.Role
	}

	roleAdmin := portaineree.Role{
		Name:           "Environment administrator",
		Description:    "Full control of all resources in an environment",
		ID:             portaineree.RoleIDEndpointAdmin,
		Priority:       1,
		Authorizations: DefaultEndpointAuthorizationsForEndpointAdministratorRole(),
	}

	roleOperator := portaineree.Role{
		Name:           "Operator",
		Description:    "Operational control of all existing resources in an environment",
		ID:             portaineree.RoleIDOperator,
		Priority:       2,
		Authorizations: DefaultEndpointAuthorizationsForOperatorRole(),
	}

	roleHelpdesk := portaineree.Role{
		Name:           "Helpdesk",
		Description:    "Read-only access of all resources in an environment",
		ID:             portaineree.RoleIDHelpdesk,
		Priority:       3,
		Authorizations: DefaultEndpointAuthorizationsForHelpDeskRole(),
	}

	roleStandard := portaineree.Role{
		Name:           "Standard user",
		Description:    "Full control of assigned resources in an environment",
		ID:             portaineree.RoleIDStandardUser,
		Priority:       4,
		Authorizations: DefaultEndpointAuthorizationsForStandardUserRole(),
	}

	roleReadonly := portaineree.Role{
		Name:           "Read-only user",
		Description:    "Read-only access of assigned resources in an environment",
		ID:             portaineree.RoleIDReadonly,
		Priority:       5,
		Authorizations: DefaultEndpointAuthorizationsForReadOnlyUserRole(),
	}

	roles := []portaineree.Role{roleAdmin, roleOperator, roleHelpdesk, roleReadonly, roleStandard}

	tests := []struct {
		name string
		args args
		want *portaineree.Role
	}{
		{
			name: "it should return Operator when Operator is before EndpointAdmin in the argument roleIdentifiers",
			args: args{
				roleIdentifiers: []portaineree.RoleID{portaineree.RoleIDOperator, portaineree.RoleIDEndpointAdmin},
				roles:           roles,
			},
			want: &roleOperator,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getKeyRole(tt.args.roleIdentifiers, tt.args.roles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getKeyRole() = %v, want %v", got, tt.want)
			}
		})
	}
}
