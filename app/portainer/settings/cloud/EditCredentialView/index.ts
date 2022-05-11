import angular from 'angular';

import { EditCredentialViewAngular } from './EditCredentialView';

export default angular
  .module('portainer.settings.cloud.credential', [])
  .component('editCloudCredentialView', EditCredentialViewAngular).name;
