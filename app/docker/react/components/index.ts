import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { StackImageStatus } from '@/react/docker/stacks/ListView/StackImageStatus';

export const componentsModule = angular
  .module('portainer.docker.react.components', [])
  .component('stackImageStatus', r2a(StackImageStatus, ['stackId'])).name;
