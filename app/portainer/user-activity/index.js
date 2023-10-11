import angular from 'angular';

import authLogsViewModule from './auth-logs-view';
import activityLogsViewModule from './activity-logs-view';

import { UserActivity } from './user-activity.rest';
import { UserActivityService } from './user-activity.service';

export default angular
  .module('portainer.app.user-activity', [authLogsViewModule, activityLogsViewModule])

  .service('UserActivity', UserActivity)
  .service('UserActivityService', UserActivityService)

  .config(config).name;

/* @ngInject */
function config($stateRegistryProvider) {
  $stateRegistryProvider.register({
    name: 'portainer.authLogs',
    url: '/auth-logs',
    views: {
      'content@': {
        component: 'authLogsView',
      },
    },
    data: {
      docs: '/admin/logs',
    },
  });

  $stateRegistryProvider.register({
    name: 'portainer.activityLogs',
    url: '/activity-logs',
    views: {
      'content@': {
        component: 'activityLogsView',
      },
    },
    data: {
      docs: '/admin/logs/activity',
    },
  });

  $stateRegistryProvider.register({
    name: 'portainer.notifications',
    url: '/notifications',
    views: {
      'content@': {
        component: 'notifications',
      },
    },
    params: {
      id: '',
    },
    data: {
      docs: '/admin/notifications',
    },
  });
}
