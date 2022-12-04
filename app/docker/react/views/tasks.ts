import { StateRegistry } from '@uirouter/angularjs';
import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { LogView } from '@/react/docker/tasks/LogView';

export const tasksModule = angular
  .module('portainer.docker.tasks', [])
  .component(
    'taskLogView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(LogView))), [])
  )
  .config(config).name;

/* @ngInject */
function config($stateRegistryProvider: StateRegistry) {
  $stateRegistryProvider.register({
    name: 'docker.tasks.task.logs',
    url: '/logs',
    views: {
      'content@': {
        component: 'taskLogView',
      },
    },
  });
}
