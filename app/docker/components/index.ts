import angular from 'angular';

import { ContainerQuickActionsAngular } from './container-quick-actions';
import ImageStatusModule from './ImageStatus';

export const componentsModule = angular
  .module('portainer.docker.components', [ImageStatusModule])
  .component('containerQuickActions', ContainerQuickActionsAngular).name;
