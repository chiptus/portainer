import { StateRegistry } from '@uirouter/angularjs';
import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { LogView } from '@/react/docker/services/LogsView';

export const servicesModule = angular
  .module('portainer.docker.services', [])
  .component(
    'serviceLogView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(LogView))), [])
  )
  .config(config).name;

/* @ngInject */
function config($stateRegistryProvider: StateRegistry) {
  $stateRegistryProvider.register({
    name: 'docker.services.service.logs',
    url: '/logs',
    views: {
      'content@': {
        component: 'serviceLogView',
      },
    },
  });
}
