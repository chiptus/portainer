import { getEnvironments } from '@/react/portainer/environments/environment.service';

import { featureFlagsModule } from './feature-flags';
import './rbac';
import './registry-management';
import licenseManagementModule from './license-management';
import settingsModule from './settings';
import userActivityModule from './user-activity';
import componentsModule from './components';
import servicesModule from './services';
import { reactModule } from './react';
import { sidebarModule } from './react/views/sidebar';
import { gitCredentialsModule } from './react/views/account/git-credentials';
import environmentsModule from './environments';
import { helpersModule } from './helpers';

async function initAuthentication(Authentication) {
  return await Authentication.init();
}

angular
  .module('portainer.app', [
    'portainer.oauth',
    'portainer.rbac',
    'portainer.registrymanagement',
    licenseManagementModule,
    componentsModule,
    settingsModule,
    userActivityModule,
    featureFlagsModule,
    'portainer.shared.datatable',
    servicesModule,
    reactModule,
    sidebarModule,
    gitCredentialsModule,
    environmentsModule,
    helpersModule,
  ])
  .config([
    '$stateRegistryProvider',
    function ($stateRegistryProvider) {
      'use strict';

      var root = {
        name: 'root',
        abstract: true,
        onEnter: /* @ngInject */ function onEnter($async, StateManager, Authentication, Notifications, $state) {
          return $async(async () => {
            const appState = StateManager.getState();
            if (!appState.loading) {
              return;
            }
            try {
              const loggedIn = await initAuthentication(Authentication);
              await StateManager.initialize();
              if (!loggedIn && isTransitionRequiresAuthentication($state.transition)) {
                $state.go('portainer.logout');
                return Promise.reject('Unauthenticated');
              }
            } catch (err) {
              Notifications.error('Failure', err, 'Unable to retrieve application settings');
              throw err;
            }
          });
        },
        views: {
          'sidebar@': {
            component: 'sidebar',
          },
        },
      };

      var endpointRoot = {
        name: 'endpoint',
        url: '/:endpointId',
        parent: 'root',
        abstract: true,
        resolve: {
          endpoint: /* @ngInject */ function endpoint($async, $state, $transition$, EndpointProvider, EndpointService, Notifications) {
            return $async(async () => {
              try {
                const endpointId = +$transition$.params().endpointId;

                const endpoint = await EndpointService.endpoint(endpointId);
                if ((endpoint.Type === 4 || endpoint.Type === 7) && !endpoint.EdgeID) {
                  $state.go('portainer.endpoints.endpoint', { id: endpoint.Id });
                  return;
                }

                EndpointProvider.setCurrentEndpoint(endpoint);

                return endpoint;
              } catch (e) {
                const nextTransition = $state.transition && $state.transition.to();
                const nextTransitionName = nextTransition ? nextTransition.name : '';
                if (nextTransitionName === 'kubernetes.nodeshell') {
                  return;
                }
                Notifications.error('Failed loading environment', e);
                $state.go('portainer.home', {}, { reload: true });
                return;
              }
            });
          },
        },
      };

      var portainer = {
        name: 'portainer',
        parent: 'root',
        abstract: true,
      };

      var account = {
        name: 'portainer.account',
        url: '/account',
        views: {
          'content@': {
            templateUrl: './views/account/account.html',
            controller: 'AccountController',
          },
        },
        data: {
          docs: '/user/account-settings',
        },
      };

      const tokenCreation = {
        name: 'portainer.account.new-access-token',
        url: '/tokens/new',
        views: {
          'content@': {
            component: 'createUserAccessToken',
          },
        },
      };

      var authentication = {
        name: 'portainer.auth',
        url: '/auth',
        params: {
          reload: false,
        },
        views: {
          'content@': {
            templateUrl: './views/auth/auth.html',
            controller: 'AuthenticationController',
            controllerAs: 'ctrl',
          },
          'sidebar@': {},
        },
      };
      var internalAuthentication = {
        name: 'portainer.internal-auth',
        url: '/internal-auth',
        params: {
          reload: false,
        },
        views: {
          'content@': {
            templateUrl: './views/internal-auth/internal-auth.html',
            controller: 'InternalAuthenticationController',
            controllerAs: 'ctrl',
          },
          'sidebar@': {},
        },
      };
      const logout = {
        name: 'portainer.logout',
        url: '/logout',
        params: {
          error: '',
        },
        views: {
          'content@': {
            templateUrl: './views/logout/logout.html',
            controller: 'LogoutController',
            controllerAs: 'ctrl',
          },
          'sidebar@': {},
        },
      };

      var endpoints = {
        name: 'portainer.endpoints',
        url: '/endpoints',
        views: {
          'content@': {
            component: 'environmentsListView',
          },
        },
        data: {
          docs: '/admin/environments',
        },
      };

      var endpoint = {
        name: 'portainer.endpoints.endpoint',
        url: '/:id?redirectTo',
        params: {
          redirectTo: '',
        },
        views: {
          'content@': {
            templateUrl: './views/endpoints/edit/endpoint.html',
            controller: 'EndpointController',
          },
        },
      };

      var deviceImport = {
        name: 'portainer.endpoints.importDevice',
        url: '/device',
        views: {
          'content@': {
            templateUrl: './views/devices/import/importDevice.html',
            controller: 'ImportDeviceController',
          },
        },
      };

      const edgeAutoCreateScript = {
        name: 'portainer.endpoints.edgeAutoCreateScript',
        url: '/aeec',
        views: {
          'content@': {
            component: 'edgeAutoCreateScriptView',
          },
        },
      };

      var addFDOProfile = {
        name: 'portainer.endpoints.profile',
        url: '/profile',
        views: {
          'content@': {
            component: 'addProfileView',
          },
        },
      };

      var editFDOProfile = {
        name: 'portainer.endpoints.profile.edit',
        url: '/:id',
        views: {
          'content@': {
            component: 'editProfileView',
          },
        },
      };

      var endpointAccess = {
        name: 'portainer.endpoints.endpoint.access',
        url: '/access',
        views: {
          'content@': {
            templateUrl: './views/endpoints/access/endpointAccess.html',
            controller: 'EndpointAccessController',
            controllerAs: 'ctrl',
          },
        },
      };

      var endpointKVM = {
        name: 'portainer.endpoints.endpoint.kvm',
        url: '/kvm?deviceId&deviceName',
        views: {
          'content@': {
            templateUrl: './views/endpoints/kvm/endpointKVM.html',
            controller: 'EndpointKVMController',
          },
        },
      };

      var groups = {
        name: 'portainer.groups',
        url: '/groups',
        views: {
          'content@': {
            templateUrl: './views/groups/groups.html',
            controller: 'GroupsController',
          },
        },
        data: {
          docs: '/admin/environments/groups',
        },
      };

      var group = {
        name: 'portainer.groups.group',
        url: '/:id',
        views: {
          'content@': {
            templateUrl: './views/groups/edit/group.html',
            controller: 'GroupController',
          },
        },
      };

      var groupCreation = {
        name: 'portainer.groups.new',
        url: '/new',
        views: {
          'content@': {
            templateUrl: './views/groups/create/creategroup.html',
            controller: 'CreateGroupController',
          },
        },
      };

      var groupAccess = {
        name: 'portainer.groups.group.access',
        url: '/access',
        views: {
          'content@': {
            templateUrl: './views/groups/access/groupAccess.html',
            controller: 'GroupAccessController',
          },
        },
      };

      var home = {
        name: 'portainer.home',
        url: '/home?redirect&environmentId&environmentName&route',
        views: {
          'content@': {
            component: 'homeView',
          },
        },
        data: {
          docs: '/user/home',
        },
      };

      var init = {
        name: 'portainer.init',
        abstract: true,
        url: '/init',
        views: {
          'sidebar@': {},
        },
      };

      var initAdmin = {
        name: 'portainer.init.admin',
        url: '/admin',
        views: {
          'content@': {
            templateUrl: './views/init/admin/initAdmin.html',
            controller: 'InitAdminController',
          },
        },
      };

      const initLicense = {
        name: 'portainer.init.license',
        url: '/license',
        views: {
          'content@': {
            component: 'initLicenseView',
          },
        },
      };

      var registries = {
        name: 'portainer.registries',
        url: '/registries',
        views: {
          'content@': {
            templateUrl: './views/registries/registries.html',
            controller: 'RegistriesController',
          },
        },
        data: {
          docs: '/admin/registries',
        },
      };

      var registry = {
        name: 'portainer.registries.registry',
        url: '/:id',
        views: {
          'content@': {
            component: 'editRegistry',
          },
        },
      };

      const registryCreation = {
        name: 'portainer.registries.new',
        url: '/new',
        views: {
          'content@': {
            component: 'createRegistry',
          },
        },
      };

      var settings = {
        name: 'portainer.settings',
        url: '/settings',
        views: {
          'content@': {
            component: 'settingsView',
          },
        },
        data: {
          docs: '/admin/settings',
        },
      };

      var settingsAuthentication = {
        name: 'portainer.settings.authentication',
        url: '/auth',
        views: {
          'content@': {
            templateUrl: './views/settings/authentication/settingsAuthentication.html',
            controller: 'SettingsAuthenticationController',
          },
        },
        data: {
          docs: '/admin/settings/authentication',
        },
      };

      const settingsCloud = {
        name: 'portainer.settings.sharedcredentials',
        url: '/cloud',
        views: {
          'content@': {
            component: 'settingsSharedCredentialsView',
          },
        },
        data: {
          docs: '/admin/settings/credentials',
        },
      };

      const addCloudCredential = {
        name: 'portainer.settings.sharedcredentials.addCredential',
        url: '/credentials/new',
        views: {
          'content@': {
            component: 'addSharedCredentialsView',
          },
        },
      };

      const editCloudCredential = {
        name: 'portainer.settings.sharedcredentials.credential',
        url: '/credentials/:id',
        views: {
          'content@': {
            component: 'editSharedCredentialsView',
          },
        },
      };

      const createGitCredential = {
        name: 'portainer.account.gitCreateGitCredential',
        url: '/git-credential/new',
        views: {
          'content@': {
            component: 'createGitCredentialView',
          },
        },
      };

      const editGitCredential = {
        name: 'portainer.account.gitEditGitCredential',
        url: '/git-credential/:id',
        views: {
          'content@': {
            component: 'editGitCredentialView',
          },
        },
      };

      const createHelmRepository = {
        name: 'portainer.account.createHelmRepository',
        url: '/helm-repository/new',
        views: {
          'content@': {
            component: 'createHelmRepositoryView',
          },
        },
      };

      var settingsEdgeCompute = {
        name: 'portainer.settings.edgeCompute',
        url: '/edge',
        views: {
          'content@': {
            component: 'settingsEdgeComputeView',
          },
        },
        data: {
          docs: '/admin/settings/edge',
        },
      };

      var tags = {
        name: 'portainer.tags',
        url: '/tags',
        views: {
          'content@': {
            templateUrl: './views/tags/tags.html',
            controller: 'TagsController',
          },
        },
        data: {
          docs: '/admin/environments/tags',
        },
      };

      var users = {
        name: 'portainer.users',
        url: '/users',
        views: {
          'content@': {
            templateUrl: './views/users/users.html',
            controller: 'UsersController',
          },
        },
        data: {
          docs: '/admin/users',
        },
      };

      var user = {
        name: 'portainer.users.user',
        url: '/:id',
        views: {
          'content@': {
            templateUrl: './views/users/edit/user.html',
            controller: 'UserController',
          },
        },
      };

      $stateRegistryProvider.register(root);
      $stateRegistryProvider.register(endpointRoot);
      $stateRegistryProvider.register(portainer);
      $stateRegistryProvider.register(account);
      $stateRegistryProvider.register(tokenCreation);
      $stateRegistryProvider.register(authentication);
      $stateRegistryProvider.register(internalAuthentication);
      $stateRegistryProvider.register(logout);
      $stateRegistryProvider.register(endpoints);
      $stateRegistryProvider.register(endpoint);
      $stateRegistryProvider.register(endpointAccess);
      $stateRegistryProvider.register(endpointKVM);
      $stateRegistryProvider.register(edgeAutoCreateScript);
      $stateRegistryProvider.register(deviceImport);
      $stateRegistryProvider.register(addFDOProfile);
      $stateRegistryProvider.register(editFDOProfile);
      $stateRegistryProvider.register(groups);
      $stateRegistryProvider.register(group);
      $stateRegistryProvider.register(groupAccess);
      $stateRegistryProvider.register(groupCreation);
      $stateRegistryProvider.register(home);
      $stateRegistryProvider.register(init);
      $stateRegistryProvider.register(initAdmin);
      $stateRegistryProvider.register(initLicense);
      $stateRegistryProvider.register(registries);
      $stateRegistryProvider.register(registry);
      $stateRegistryProvider.register(registryCreation);
      $stateRegistryProvider.register(settings);
      $stateRegistryProvider.register(settingsAuthentication);
      $stateRegistryProvider.register(settingsCloud);
      $stateRegistryProvider.register(addCloudCredential);
      $stateRegistryProvider.register(editCloudCredential);
      $stateRegistryProvider.register(createGitCredential);
      $stateRegistryProvider.register(editGitCredential);
      $stateRegistryProvider.register(settingsEdgeCompute);
      $stateRegistryProvider.register(tags);
      $stateRegistryProvider.register(users);
      $stateRegistryProvider.register(user);
      $stateRegistryProvider.register(createHelmRepository);
    },
  ])
  .run(run);

const UNAUTHENTICATED_ROUTES = ['portainer.logout', 'portainer.internal-auth', 'portainer.auth', 'portainer.init.admin'];
function isTransitionRequiresAuthentication(transition) {
  if (!transition) {
    return true;
  }
  const nextTransition = transition && transition.to();
  const nextTransitionName = nextTransition ? nextTransition.name : '';
  return !UNAUTHENTICATED_ROUTES.some((route) => nextTransitionName.startsWith(route));
}

/* @ngInject */
function run($transitions, UserService, Authentication, LicenseService, Notifications) {
  $transitions.onBefore({ to: 'portainer.init.*' }, async function (transition) {
    const to = transition.to();
    const stateService = transition.router.stateService;

    try {
      const adminExists = await UserService.administratorExists();
      if (!adminExists) {
        return to.name !== 'portainer.init.admin' ? stateService.target('portainer.init.admin') : true;
      }
    } catch (err) {
      Notifications.error('Failure', err, 'Unable to retrieve admin');
      throw err;
    }

    if (!Authentication.isAuthenticated()) {
      return stateService.target('portainer.auth');
    }

    try {
      const licenseInfo = await LicenseService.info();
      if (!licenseInfo.valid) {
        return to.name !== 'portainer.init.license' ? stateService.target('portainer.init.license') : true;
      }
    } catch (err) {
      Notifications.error('Failure', err, 'Unable to retrieve license info');
      throw err;
    }

    try {
      const endpoints = await getEnvironments({ start: 0, limit: 1, query: { excludeSnapshots: true } });
      if (endpoints.value.length === 0) {
        return to.name !== 'portainer.wizard' ? stateService.target('portainer.wizard') : true;
      }

      return stateService.target('portainer.home');
    } catch (err) {
      Notifications.error('Failure', err, 'Unable to retrieve environment info');
      throw err;
    }
  });

  $transitions.onBefore({ to: (state) => !state.name.startsWith('portainer.init') && !UNAUTHENTICATED_ROUTES.includes(state.name) }, function (transition) {
    const stateService = transition.router.stateService;

    async function licenseCheckAsync() {
      try {
        const licenseInfo = await LicenseService.info();
        if (!licenseInfo.valid) {
          return stateService.target('portainer.init.license');
        }
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to retrieve license info');
        throw err;
      }
    }

    licenseCheckAsync();
  });
}
