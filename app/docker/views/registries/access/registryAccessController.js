import { TeamAccessViewModel, UserAccessViewModel } from 'Portainer/models/access';

class DockerRegistryAccessController {
  /* @ngInject */
  constructor($async, $state, Authentication, Notifications, EndpointService, GroupService, RoleService) {
    this.$async = $async;
    this.$state = $state;
    this.Authentication = Authentication;
    this.Notifications = Notifications;
    this.EndpointService = EndpointService;
    this.GroupService = GroupService;
    this.RoleService = RoleService;

    this.updateAccess = this.updateAccess.bind(this);
    this.filterUsers = this.filterUsers.bind(this);
  }

  updateAccess() {
    return this.$async(async () => {
      this.state.actionInProgress = true;
      try {
        await this.EndpointService.updateRegistryAccess(this.state.endpointId, this.state.registryId, this.registryEndpointAccesses);
        this.Notifications.success('Access successfully updated');
        this.$state.reload();
      } catch (err) {
        this.state.actionInProgress = false;
        this.Notifications.error('Failure', err, 'Unable to update accesses');
      }
    });
  }

  filterUsers(users) {
    const endpointUsers = this.endpoint.UserAccessPolicies;
    const endpointTeams = this.endpoint.TeamAccessPolicies;

    const endpointGroupUsers = this.endpointGroup.UserAccessPolicies;
    const endpointGroupTeams = this.endpointGroup.TeamAccessPolicies;

    return users.filter((userOrTeam) => {
      const userAccess = userOrTeam instanceof UserAccessViewModel && (endpointUsers[userOrTeam.Id] || endpointGroupUsers[userOrTeam.Id]);
      const teamAccess = userOrTeam instanceof TeamAccessViewModel && (endpointTeams[userOrTeam.Id] || endpointGroupTeams[userOrTeam.Id]);

      const entityAccess = userAccess || teamAccess;
      if (!entityAccess) {
        return false;
      }

      const roleId = entityAccess.RoleId;
      const role = this.roles.find((role) => role.Id === roleId);
      return this.isStandardRole(role);
    });
  }

  isStandardRole(role) {
    return role && (role.Authorizations['DockerImageCreate'] || role.Authorizations['DockerImagePush']) && !role.Authorizations['EndpointResourcesAccess'];
  }

  $onInit() {
    return this.$async(async () => {
      this.Authentication.redirectIfUnauthorized(['OperationPortainerRegistryUpdateAccess']);

      try {
        this.state = {
          viewReady: false,
          actionInProgress: false,
          endpointId: this.$state.params.endpointId,
          registryId: this.$state.params.id,
        };
        this.roles = await this.RoleService.roles();
        this.registry = await this.EndpointService.registry(this.state.endpointId, this.state.registryId);
        this.registryEndpointAccesses = this.registry.RegistryAccesses[this.state.endpointId] || {};
        this.endpointGroup = await this.GroupService.group(this.endpoint.GroupId);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve registry details');
      } finally {
        this.state.viewReady = true;
      }
    });
  }
}

export default DockerRegistryAccessController;
angular.module('portainer.docker').controller('DockerRegistryAccessController', DockerRegistryAccessController);
