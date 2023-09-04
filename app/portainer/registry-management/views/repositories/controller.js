import _ from 'lodash-es';

import { RegistryTypes } from 'Portainer/models/registryTypes';
import EndpointHelper from '@/portainer/helpers/endpointHelper';
import { getInfo } from '@/react/docker/proxy/queries/useInfo';

export class RegistryRepositoriesController {
  /* @ngInject */
  constructor($async, $state, EndpointService, RegistryService, RegistryServiceSelector, Notifications, Authentication) {
    Object.assign(this, { $async, $state, EndpointService, RegistryService, RegistryServiceSelector, Notifications, Authentication });

    this.state = {
      displayInvalidConfigurationMessage: false,
      loading: false,
    };

    this.paginationAction = this.paginationAction.bind(this);
    this.paginationActionAsync = this.paginationActionAsync.bind(this);
    this.endpointProviderType = this.endpointProviderType.bind(this);
    this.$onInit = this.$onInit.bind(this);
  }

  getRegistriesLink() {
    switch (this.endpointProviderType) {
      case 'swarm':
        return 'docker.swarm.registries';
      case 'docker':
        return 'docker.host.registries';
      case 'kubernetes':
        return 'kubernetes.registries';
      default:
        return 'portainer.registries';
    }
  }

  paginationAction(repositories) {
    return this.$async(this.paginationActionAsync, repositories);
  }
  async paginationActionAsync(repositories) {
    if (this.registry.Type === RegistryTypes.GITLAB) {
      return;
    }
    this.state.loading = true;
    try {
      const data = await this.RegistryServiceSelector.getRepositoriesDetails(this.registry, this.endpointId, repositories);
      for (let i = 0; i < data.length; i++) {
        const idx = _.findIndex(this.repositories, { Name: data[i].Name });
        if (idx !== -1) {
          if (data[i].TagsCount === 0) {
            this.repositories.splice(idx, 1);
          } else {
            this.repositories[idx].TagsCount = data[i].TagsCount;
          }
        }
      }
      this.state.loading = false;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve repositories details');
    }
  }

  // decide the endpoint provider type to decide the registry list link
  async endpointProviderType(endpointId) {
    // if the endpoint is not in the query params, then we are in the main registries view
    if (!endpointId) {
      return '';
    }

    // otherwise return the environment provider type
    const endpoint = await this.EndpointService.endpoint(endpointId);
    const isDockerOrSwarmEndpoint = EndpointHelper.isDockerEndpoint(endpoint);
    if (isDockerOrSwarmEndpoint) {
      const endpointInfo = await getInfo(endpoint.Id);
      if (endpointInfo.Swarm.NodeID) {
        return 'swarm';
      }
      return 'docker';
    }
    if (EndpointHelper.isKubernetesEndpoint(endpoint)) {
      return 'kubernetes';
    }
    return '';
  }

  async $onInit() {
    const registryId = this.$state.params.id;

    this.isAdmin = this.Authentication.isAdmin();
    this.endpointId = this.$state.params.endpointId;
    this.endpointProviderType = await this.endpointProviderType(this.endpointId);

    try {
      this.registry = await this.RegistryService.registry(registryId, this.endpointId);
      try {
        await this.RegistryServiceSelector.ping(this.registry, this.endpointId, false);
        this.repositories = await this.RegistryServiceSelector.repositories(this.registry, this.endpointId);
      } catch (e) {
        this.state.displayInvalidConfigurationMessage = true;
      }
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve registry details');
    }
  }
}
