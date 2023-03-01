import _ from 'lodash-es';
import { RegistryImageLayerViewModel } from '@/portainer/registry-management/models/registryImageLayer';
import { RegistryImageDetailsViewModel } from '@/portainer/registry-management/models/registryImageDetails';
import EndpointHelper from '@/portainer/helpers/endpointHelper';
import { getInfo } from '@/docker/services/system.service';

export class RegistryRepositoryTagController {
  /* @ngInject */
  constructor($state, $async, Notifications, RegistryService, RegistryServiceSelector, imagelayercommandFilter, EndpointService) {
    Object.assign(this, { $state, $async, Notifications, RegistryService, RegistryServiceSelector, imagelayercommandFilter, EndpointService });

    this.context = {};

    this.$onInit = this.$onInit.bind(this);
    this.$onInitAsync = this.$onInitAsync.bind(this);
    this.getEndpointProviderType = this.getEndpointProviderType.bind(this);
    this.getRegistriesLink = this.getRegistriesLink.bind(this);
  }

  toggleLayerCommand(layerId) {
    $('#layer-command-expander' + layerId + ' span').toggleClass('glyphicon-plus-sign glyphicon-minus-sign');
    $('#layer-command-' + layerId + '-short').toggle();
    $('#layer-command-' + layerId + '-full').toggle();
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

  async getEndpointProviderType(endpointId) {
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

  order(sortType) {
    this.Sort.Reverse = this.Sort.Type === sortType ? !this.Sort.Reverse : false;
    this.Sort.Type = sortType;
  }

  $onInit() {
    return this.$async(this.$onInitAsync);
  }
  async $onInitAsync() {
    this.context.registryId = this.$state.params.id;
    this.context.repository = this.$state.params.repository;
    this.context.tag = this.$state.params.tag;
    this.context.endpointId = this.$state.params.endpointId;
    this.endpointProviderType = await this.getEndpointProviderType(this.context.endpointId);
    this.Sort = {
      Type: 'Order',
      Reverse: false,
    };
    try {
      this.registry = await this.RegistryService.registry(this.context.registryId, this.context.endpointId);
      this.tag = await this.RegistryServiceSelector.tag(this.registry, this.context.endpointId, this.context.repository, this.context.tag);
      const length = this.tag.History.length;
      this.history = _.map(this.tag.History, (layer, idx) => new RegistryImageLayerViewModel(length - idx, layer));
      _.forEach(this.history, (item) => (item.CreatedBy = this.imagelayercommandFilter(item.CreatedBy)));
      this.details = new RegistryImageDetailsViewModel(this.tag.History[0]);
    } catch (error) {
      this.Notifications.error('Failure', error, 'Unable to retrieve tag');
    }
  }
}
