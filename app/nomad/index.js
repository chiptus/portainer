import { isNomadEnvironment } from '@/portainer/environments/utils';
import { NomadDashboardAngular } from './Dashboard';
import { JobsViewAngular } from './Jobs';
import { NomadSidebarAngular } from './NomadSidebar';
import { NomadEventsAngular } from './Events/Events';
import { reactModule } from './react';

function config($stateRegistryProvider) {
  'use strict';

  const nomad = {
    name: 'nomad',
    url: '/nomad',
    parent: 'endpoint',
    abstract: true,

    onEnter: /* @ngInject */ function onEnter($async, $state, endpoint, Notifications, StateManager, EndpointProvider) {
      return $async(async () => {
        if (!isNomadEnvironment(endpoint.Type)) {
          $state.go('portainer.home');
          return;
        }
        try {
          EndpointProvider.setEndpointID(endpoint.Id);
          await StateManager.updateEndpointState(endpoint, []);
        } catch (e) {
          Notifications.error('Failed loading environment', e);
          $state.go('portainer.home', {}, { reload: true });
        }
      });
    },
  };

  const dashboard = {
    name: 'nomad.dashboard',
    url: '/dashboard',
    views: {
      'content@': {
        component: 'nomadDashboard',
      },
    },
  };

  const jobs = {
    name: 'nomad.jobs',
    url: '/jobs',
    views: {
      'content@': {
        component: 'nomadJobs',
      },
    },
  };

  const events = {
    name: 'nomad.events',
    url: '/jobs/:jobID/tasks/:taskName/allocations/:allocationID/events?namespace',
    views: {
      'content@': {
        component: 'nomadEvents',
      },
    },
  };

  const logs = {
    name: 'nomad.logs',
    url: '/jobs/:jobID/tasks/:taskName/allocations/:allocationID/logs?namespace',
    views: {
      'content@': {
        templateUrl: './Logs/logs.html',
        controller: 'LogsController',
      },
    },
  };

  $stateRegistryProvider.register(nomad);
  $stateRegistryProvider.register(dashboard);
  $stateRegistryProvider.register(jobs);
  $stateRegistryProvider.register(events);
  $stateRegistryProvider.register(logs);
}

export const nomadModule = angular
  .module('portainer.nomad', ['portainer.app', reactModule])
  .config(['$stateRegistryProvider', config])
  .component('nomadDashboard', NomadDashboardAngular)
  .component('nomadSidebar', NomadSidebarAngular)
  .component('nomadEvents', NomadEventsAngular)
  .component('nomadJobs', JobsViewAngular).name;
