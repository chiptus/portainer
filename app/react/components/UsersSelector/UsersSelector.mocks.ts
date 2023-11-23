import { UserViewModel } from '@/portainer/models/user';
import { Role, RoleNames } from '@/portainer/users/types';

export function createMockUser(id: number, username: string): UserViewModel {
  return {
    Id: id,
    Username: username,
    Role: Role.Standard,
    EndpointAuthorizations: {},
    UseCache: false,
    PortainerAuthorizations: {
      PortainerDockerHubInspect: true,
      PortainerEndpointGroupInspect: true,
      PortainerEndpointGroupList: true,
      PortainerEndpointInspect: true,
      PortainerEndpointList: true,
      PortainerMOTD: true,
      PortainerRoleList: true,
      PortainerTeamList: true,
      PortainerTemplateInspect: true,
      PortainerTemplateList: true,
      PortainerUserInspect: true,
      PortainerUserList: true,
      PortainerUserMemberships: true,
    },
    RoleName: RoleNames[Role.Standard],
    Checked: false,
    AuthenticationMethod: '',
    ThemeSettings: {
      color: 'auto',
      subtleUpgradeButton: false,
    },
  };
}
