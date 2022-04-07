import _ from 'lodash-es';
import { RegistryTypes } from 'Portainer/models/registryTypes';

class EndpointRegistriesController {
  /* @ngInject */
  constructor($async, Authentication, Notifications, EndpointService) {
    this.$async = $async;
    this.Authentication = Authentication;
    this.Notifications = Notifications;
    this.EndpointService = EndpointService;

    this.canManageAccess = this.canManageAccess.bind(this);
    this.canBrowse = this.canBrowse.bind(this);
  }

  canManageAccess(item) {
    return item.Type !== RegistryTypes.ANONYMOUS && this.Authentication.hasAuthorizations(['PortainerRegistryUpdateAccess']);
  }

  canBrowse(item) {
    return !_.includes([RegistryTypes.ANONYMOUS, RegistryTypes.DOCKERHUB, RegistryTypes.QUAY], item.Type);
  }

  getRegistries() {
    return this.$async(async () => {
      try {
        this.registries = await this.EndpointService.registries(this.endpointId);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve registries');
      }
    });
  }

  $onInit() {
    return this.$async(async () => {
      this.state = {
        viewReady: false,
      };

      try {
        this.endpointType = this.endpoint.Type;
        this.endpointId = this.endpoint.Id;
        await this.getRegistries();
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve registries');
      } finally {
        this.state.viewReady = true;
      }
    });
  }
}

export default EndpointRegistriesController;
angular.module('portainer.app').controller('EndpointRegistriesController', EndpointRegistriesController);
