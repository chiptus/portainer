import angular from 'angular';

import kaasModule from './kaas';
import { azureEndpointConfig } from './azure-endpoint-config/azure-endpoint-config';

export default angular
  .module('portainer.environments', [kaasModule])
  .component('azureEndpointConfig', azureEndpointConfig).name;
