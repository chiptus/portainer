import angular from 'angular';

import { CloudViewAngular } from './CloudView';

export default angular
  .module('portainer.settings.cloud', [])
  .component('settingsCloudView', CloudViewAngular).name;
