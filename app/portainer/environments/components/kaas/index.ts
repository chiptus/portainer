import angular from 'angular';

import { KaasCreateFormAngular } from './KaasCreate.form';

export default angular
  .module('portainer.environments.components.kaas', [])
  .component('kaasCreateForm', KaasCreateFormAngular).name;
