import { TeamViewModel } from '@/portainer/models/team';
import { UserViewModel } from '@/portainer/models/user';
import { Role, RoleNames } from '@/portainer/users/types';

export function mockExampleData() {
  const teams: TeamViewModel[] = [
    {
      Id: 3,
      Name: 'Team 1',
      Checked: false,
    },
    {
      Id: 4,
      Name: 'Team 2',
      Checked: false,
    },
  ];

  const users: UserViewModel[] = [
    {
      Id: 10,
      Username: 'user1',
      Role: Role.Standard,
      ThemeSettings: {
        color: 'auto',
        subtleUpgradeButton: false,
      },
      EndpointAuthorizations: {},
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
      UseCache: false,
    },
    {
      Id: 13,
      Username: 'user2',
      Role: Role.Standard,
      ThemeSettings: {
        color: 'auto',
        subtleUpgradeButton: false,
      },
      EndpointAuthorizations: {},
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
      UseCache: false,
    },
  ];

  return { users, teams };
}
