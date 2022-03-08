import angular from 'angular';

import { LicenseService } from './license.service.ts';
import licensesViewModule from './licenses.view';
import addLicenseViewModule from './add-license.view';

export default angular.module('portainer.app.license-management', [licensesViewModule, addLicenseViewModule]).config(config).service('LicenseService', LicenseService).name;

/* @ngInject */
function config($stateRegistryProvider) {
  const licenses = {
    name: 'portainer.licenses',
    url: '/licenses',
    views: {
      'content@': {
        component: 'licensesView',
      },
    },
    onEnter: /* @ngInject */ function onEnter($async, $state, Authentication) {
      return $async(async () => {
        if (!Authentication.isAdmin()) {
          return $state.go('portainer.home');
        }
      });
    },
  };

  const addLicense = {
    name: 'portainer.licenses.new',
    url: '/licenses/new',
    views: {
      'content@': {
        component: 'addLicenseView',
      },
    },
  };

  $stateRegistryProvider.register(licenses);
  $stateRegistryProvider.register(addLicense);
}
