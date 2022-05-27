import angular from 'angular';

import { KaasFormGroupAngular } from './KaasFormGroup';

export default angular
  .module('portainer.environments.components.kaas', [])
  .component('kaasCreateForm', KaasFormGroupAngular).name;
