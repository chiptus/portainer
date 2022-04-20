import angular from 'angular';

import './wizard-endpoint-type.css';

angular.module('portainer.app').component('wizardEndpointType', {
  templateUrl: './wizard-endpoint-type.html',
  bindings: {
    endpointTitle: '@',
    description: '@',
    icon: '@',
    icon2: '@?',
    active: '<',
  },
});
