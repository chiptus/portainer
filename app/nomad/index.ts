import angular from 'angular';
import { StateRegistry, StateService } from '@uirouter/angularjs';

import { isNomadEnvironment } from '@/react/portainer/environments/utils';
import { DashboardView } from '@/react/nomad/DashboardView';
import { r2a } from '@/react-tools/react2angular';
import { EventsView } from '@/react/nomad/jobs/EventsView';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { JobsView } from '@/react/nomad/jobs/JobsView';
import { getLeader } from '@/react/nomad/nomad.service';
import { Environment } from '@/react/portainer/environments/types';
import { StateManager } from '@/portainer/services/types';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { LogsView } from '@/react/nomad/jobs/LogsView';

import { reactModule } from './react';

export const nomadModule = angular
  .module('portainer.nomad', [reactModule])
  .config(config)
  .component(
    'nomadLogsView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(LogsView))), [])
  )
  .component(
    'nomadDashboardView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(DashboardView))), [])
  )
  .component(
    'nomadEventsView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EventsView))), [])
  )
  .component(
    'nomadJobsView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(JobsView))), [])
  ).name;

/* @ngInject */
function config($stateRegistryProvider: StateRegistry) {
  // limits module to BE only
  if (!isBE) {
    return;
  }

  const nomad = {
    name: 'nomad',
    url: '/nomad',
    parent: 'endpoint',
    abstract: true,

    onEnter: /* @ngInject */ function onEnter(
      $async: (fn: () => Promise<void>) => Promise<void>,
      $state: StateService,
      endpoint: Environment,
      StateManager: StateManager
    ) {
      return $async(async () => {
        if (!isNomadEnvironment(endpoint.Type)) {
          $state.go('portainer.home');
          return;
        }

        try {
          await getLeader(endpoint.Id);
          await StateManager.updateEndpointState(endpoint);
        } catch (e) {
          const params = {
            redirect: true,
            environmentId: endpoint.Id,
            environmentName: endpoint.Name,
            route: 'nomad.dashboard',
          };
          $state.go('portainer.home', params, { reload: true, inherit: false });
        }
      });
    },
  };

  const dashboard = {
    name: 'nomad.dashboard',
    url: '/dashboard',
    views: {
      'content@': {
        component: 'nomadDashboardView',
      },
    },
    data: {
      docs: '/user/nomad/dashboard',
    },
  };

  const jobs = {
    name: 'nomad.jobs',
    url: '/jobs',
    views: {
      'content@': {
        component: 'nomadJobsView',
      },
    },
    data: {
      docs: '/user/nomad/jobs',
    },
  };

  const events = {
    name: 'nomad.events',
    url: '/jobs/:jobID/tasks/:taskName/allocations/:allocationID/events?namespace',
    views: {
      'content@': {
        component: 'nomadEventsView',
      },
    },
  };

  const logs = {
    name: 'nomad.logs',
    url: '/jobs/:jobID/tasks/:taskName/allocations/:allocationID/logs?namespace',
    views: {
      'content@': {
        component: 'nomadLogsView',
      },
    },
  };

  $stateRegistryProvider.register(nomad);
  $stateRegistryProvider.register(dashboard);
  $stateRegistryProvider.register(jobs);
  $stateRegistryProvider.register(events);
  $stateRegistryProvider.register(logs);
}
