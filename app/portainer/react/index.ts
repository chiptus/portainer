import angular from 'angular';

import { gitopsModule } from '@/react/portainer/gitops';

import { componentsModule } from './components';
import { viewsModule } from './views';

export const reactModule = angular.module('portainer.app.react', [
  viewsModule,
  componentsModule,
  gitopsModule,
]).name;
