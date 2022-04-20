import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';

import { WizardEndpointsList } from './WizardEndpointsList';

angular
  .module('portainer.app')
  .component('wizardEndpointList', r2a(WizardEndpointsList, ['environments']));
