import angular from 'angular';

import { CreateCredentialViewAngular } from './CreateCredentialView';

export default angular
  .module('portainer.settings.cloud.addCredential', [])
  .component('addCloudCredentialView', CreateCredentialViewAngular).name;
