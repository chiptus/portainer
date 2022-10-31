import angular from 'angular';

import controller from './serviceConfigs.controller';

angular.module('portainer.docker').component('serviceConfigs', {
  templateUrl: './configs.html',
  controller,
  bindings: {
    onSubmit: '<',
    service: '<',
    value: '<',
    onChange: '<',
    hasChanges: '<',
    cancelChanges: '<',
  },
});
