import angular from 'angular';

import { isSnapshotBrowsingSupported } from '@/react/portainer/environments/utils';
import { notifyError } from '@/portainer/services/notifications';
import { getEndpoint } from '@/react/portainer/environments/environment.service';
import edgeStackModule from './views/edge-stacks';
import { reactModule } from './react';

angular
  .module('portainer.edge', [edgeStackModule, reactModule])

  .config(function config($stateRegistryProvider) {
    const edge = {
      name: 'edge',
      url: '/edge',
      parent: 'root',
      abstract: true,
    };

    const groups = {
      name: 'edge.groups',
      url: '/groups',
      views: {
        'content@': {
          component: 'edgeGroupsView',
        },
      },
    };

    const groupsNew = {
      name: 'edge.groups.new',
      url: '/new',
      views: {
        'content@': {
          component: 'createEdgeGroupView',
        },
      },
    };

    const groupsEdit = {
      name: 'edge.groups.edit',
      url: '/:groupId',
      views: {
        'content@': {
          component: 'editEdgeGroupView',
        },
      },
    };

    const stacks = {
      name: 'edge.stacks',
      url: '/stacks',
      views: {
        'content@': {
          component: 'edgeStacksView',
        },
      },
    };

    const stacksNew = {
      name: 'edge.stacks.new',
      url: '/new',
      views: {
        'content@': {
          component: 'createEdgeStackView',
        },
      },
    };

    const stacksEdit = {
      name: 'edge.stacks.edit',
      url: '/:stackId?tab&status',
      views: {
        'content@': {
          component: 'editEdgeStackView',
        },
      },
      params: {
        status: {
          dynamic: true,
        },
      },
    };

    const edgeJobs = {
      name: 'edge.jobs',
      url: '/jobs',
      views: {
        'content@': {
          component: 'edgeJobsView',
        },
      },
    };

    const edgeJob = {
      name: 'edge.jobs.job',
      url: '/:id',
      views: {
        'content@': {
          component: 'edgeJobView',
        },
      },
      params: {
        tab: 0,
      },
    };

    const edgeJobCreation = {
      name: 'edge.jobs.new',
      url: '/new',
      views: {
        'content@': {
          component: 'createEdgeJobView',
        },
      },
    };

    $stateRegistryProvider.register({
      name: 'edge.devices',
      url: '/devices',
      abstract: true,
    });

    if (process.env.PORTAINER_EDITION === 'BE') {
      $stateRegistryProvider.register({
        name: 'edge.devices.waiting-room',
        url: '/waiting-room',
        views: {
          'content@': {
            component: 'waitingRoomView',
          },
        },
      });
    }

    $stateRegistryProvider.register({
      name: 'edge.browse',
      url: '/browse/:environmentId',
      abstract: true,
      resolve: {
        endpoint: /* @ngInject */ function endpoint($async, $transition$, $state) {
          return $async(async () => {
            try {
              const endpointId = +$transition$.params().environmentId;

              const environment = await getEndpoint(endpointId);

              if (!isSnapshotBrowsingSupported(environment)) {
                throw new Error('Snapshot browsing is not supported for this environment');
              }

              return environment;
            } catch (e) {
              notifyError('Failed loading environment', e);

              // if previous state is not navigable, go to home
              const fromState = $transition$.$from();
              if (!fromState.navigable) {
                $state.go('portainer.home');
              }
              throw e;
            }
          });
        },
      },
    });

    $stateRegistryProvider.register({
      name: 'edge.browse.dashboard',
      url: '/dashboard',
      views: {
        'content@': {
          component: 'edgeDeviceDashboardView',
        },
      },
    });

    $stateRegistryProvider.register({
      name: 'edge.browse.stacks',
      url: '/:environmentId/stacks?edgeStackId',
      views: {
        'content@': {
          component: 'edgeDeviceStacksView',
        },
      },
    });

    $stateRegistryProvider.register({
      name: 'edge.browse.containers',
      url: '/containers?edgeStackId',
      views: {
        'content@': {
          component: 'edgeDeviceContainersView',
        },
      },
    });

    $stateRegistryProvider.register({
      name: 'edge.browse.containers.container',
      url: '/:containerId',
      views: {
        'content@': {
          component: 'edgeDeviceContainerView',
        },
      },
    });

    $stateRegistryProvider.register({
      name: 'edge.browse.images',
      url: '/images',
      views: {
        'content@': {
          component: 'edgeDeviceImagesView',
        },
      },
    });

    $stateRegistryProvider.register({
      name: 'edge.browse.volumes',
      url: '/volumes',
      views: {
        'content@': {
          component: 'edgeDeviceVolumesView',
        },
      },
    });

    $stateRegistryProvider.register(edge);

    $stateRegistryProvider.register(groups);
    $stateRegistryProvider.register(groupsNew);
    $stateRegistryProvider.register(groupsEdit);

    $stateRegistryProvider.register(stacks);
    $stateRegistryProvider.register(stacksNew);
    $stateRegistryProvider.register(stacksEdit);

    $stateRegistryProvider.register(edgeJobs);
    $stateRegistryProvider.register(edgeJob);
    $stateRegistryProvider.register(edgeJobCreation);
  });
