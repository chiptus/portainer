import angular from 'angular';

import authenticationModule from './authentication';
import generalModule from './general';
import cloudModule from './cloud';

import { SettingsFDOAngular } from './edge-compute/SettingsFDO';
import { SettingsOpenAMTAngular } from './edge-compute/SettingsOpenAMT';
import { EdgeComputeSettingsViewAngular } from './edge-compute/EdgeComputeSettingsView';

export default angular
  .module('portainer.settings', [authenticationModule, generalModule, cloudModule])
  .component('settingsEdgeCompute', EdgeComputeSettingsViewAngular)
  .component('settingsFdo', SettingsFDOAngular)
  .component('settingsOpenAmt', SettingsOpenAMTAngular).name;
