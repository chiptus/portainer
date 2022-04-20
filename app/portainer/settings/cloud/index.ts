import angular from 'angular';

import { CloudViewAngular } from './CloudView';
import { CloudSettingsFormAngular } from './CloudSettingsForm';

export default angular
  .module('portainer.settings.cloud', [])
  .component('settingsCloudView', CloudViewAngular)
  .component('settingsCloudForm', CloudSettingsFormAngular).name;
