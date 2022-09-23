import angular from 'angular';

import { NotificationsViewAngular } from '@/react/portainer/notifications/NotificationsView';
import authLogsViewModule from './auth-logs-view';
import activityLogsViewModule from './activity-logs-view';

import { UserActivity } from './user-activity.rest';
import { UserActivityService } from './user-activity.service';

export default angular
  .module('portainer.app.user-activity', [authLogsViewModule, activityLogsViewModule])
  .component('notifications', NotificationsViewAngular)
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
  });

  $stateRegistryProvider.register({
    name: 'portainer.activityLogs',
    url: '/activity-logs',
    views: {
      'content@': {
        component: 'activityLogsView',
      },
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
  });
}
