import angular from 'angular';
import { StateRegistry } from '@uirouter/angularjs';

import { r2a } from '@/react-tools/react2angular';
import {
  ListView,
  CreateView,
  ItemView,
} from '@/react/edge/edge-configurations';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withCurrentUser } from '@/react-tools/withCurrentUser';

export const configurationsModule = angular
  .module('portainer.edge.configurations', [])
  .component(
    'edgeConfigurationsListView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ListView))), [])
  )
  .component(
    'edgeConfigurationsCreateView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(CreateView))), [])
  )
  .component(
    'edgeConfigurationsItemView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ItemView))), [])
  )
  .config(config).name;

function config($stateRegistryProvider: StateRegistry) {
  if (process.env.PORTAINER_EDITION === 'BE') {
    $stateRegistryProvider.register({
      name: 'edge.configurations',
      url: '/configurations',
      views: {
        'content@': {
          component: 'edgeConfigurationsListView',
        },
      },
    });

    $stateRegistryProvider.register({
      name: 'edge.configurations.create',
      url: '/new',
      views: {
        'content@': {
          component: 'edgeConfigurationsCreateView',
        },
      },
    });

    $stateRegistryProvider.register({
      name: 'edge.configurations.item',
      url: '/:id',
      views: {
        'content@': {
          component: 'edgeConfigurationsItemView',
        },
      },
    });
  }
}
