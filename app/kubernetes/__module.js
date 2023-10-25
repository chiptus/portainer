import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';
import { EnvironmentStatus } from '@/react/portainer/environments/types';
import { getSelfSubjectAccessReview } from '@/react/kubernetes/namespaces/getSelfSubjectAccessReview';

import { PortainerEndpointTypes } from 'Portainer/models/endpoint/models';

import registriesModule from './registries';
import customTemplateModule from './custom-templates';
import { reactModule } from './react';
import './views/kubernetes.css';

angular.module('portainer.kubernetes', ['portainer.app', registriesModule, customTemplateModule, reactModule]).config([
  '$stateRegistryProvider',
  function ($stateRegistryProvider) {
    'use strict';

    const kubernetes = {
      name: 'kubernetes',
      url: '/kubernetes',
      parent: 'endpoint',
      abstract: true,

      onEnter: /* @ngInject */ function onEnter($async, $state, endpoint, KubernetesHealthService, KubernetesNamespaceService, Notifications, StateManager) {
        return $async(async () => {
          const kubeTypes = [
            PortainerEndpointTypes.KubernetesLocalEnvironment,
            PortainerEndpointTypes.AgentOnKubernetesEnvironment,
            PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment,
          ];

          const nextTransition = $state.transition && $state.transition.to();
          const nextTransitionName = nextTransition ? nextTransition.name : '';
          if (nextTransitionName === 'kubernetes.nodeshell' && !endpoint) {
            return;
          }

          if (!kubeTypes.includes(endpoint.Type)) {
            $state.go('portainer.home');
            return;
          }
          try {
            if (endpoint.Type === PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment) {
              //edge
              try {
                await KubernetesHealthService.ping(endpoint.Id);
                endpoint.Status = EnvironmentStatus.Up;
              } catch (e) {
                endpoint.Status = EnvironmentStatus.Down;
              }
            }

            await StateManager.updateEndpointState(endpoint);

            if (endpoint.Type === PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment && endpoint.Status === EnvironmentStatus.Down) {
              throw new Error('Unable to contact Edge agent, please ensure that the agent is properly running on the remote environment.');
            }

            // use selfsubject access review to check if we can connect to the kubernetes environment
            // because it's gets a fast response, and is accessible to all users
            try {
              await getSelfSubjectAccessReview(endpoint.Id, 'default');
            } catch (e) {
              throw new Error(`The environment named ${endpoint.Name} is unreachable.`);
            }
          } catch (e) {
            let params = {};

            if (endpoint.Type == PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment) {
              params = { redirect: true, environmentId: endpoint.Id, environmentName: endpoint.Name, route: 'kubernetes.dashboard' };
            } else {
              Notifications.error('Failed loading environment', e);
            }
            if (nextTransitionName === 'kubernetes.nodeshell') {
              return;
            }
            $state.go('portainer.home', params, { reload: true, inherit: false });
          }
        });
      },
      views: {
        'chatBotItem@': 'chatBotItem',
      },
    };

    const helmApplication = {
      name: 'kubernetes.helm',
      url: '/helm/:namespace/:name',
      views: {
        'content@': {
          component: 'kubernetesHelmApplicationView',
        },
      },
      data: {
        docs: '/user/kubernetes/helm',
      },
    };

    const services = {
      name: 'kubernetes.services',
      url: '/services',
      views: {
        'content@': {
          component: 'kubernetesServicesView',
        },
      },
      data: {
        docs: '/user/kubernetes/services',
      },
    };

    const ingresses = {
      name: 'kubernetes.ingresses',
      url: '/ingresses',
      views: {
        'content@': {
          component: 'kubernetesIngressesView',
        },
      },
      data: {
        docs: '/user/kubernetes/ingresses',
      },
    };

    const ingressesCreate = {
      name: 'kubernetes.ingresses.create',
      url: '/add',
      views: {
        'content@': {
          component: 'kubernetesIngressesCreateView',
        },
      },
      onEnter: /* @ngInject */ function endpoint($async, $state, $transition$, Notifications) {
        return $async(async () => {
          try {
            const endpointId = +$transition$.params().endpointId;
            const deploymentOptions = await getDeploymentOptions(endpointId);
            if (deploymentOptions.hideAddWithForm) {
              $state.go('kubernetes.ingresses', { endpointId });
            }
          } catch (err) {
            Notifications.error('Failed to get deployment options', err);
          }
        });
      },
    };

    const ingressesEdit = {
      name: 'kubernetes.ingresses.edit',
      url: '/:namespace/:name/edit',
      views: {
        'content@': {
          component: 'kubernetesIngressesCreateView',
        },
      },
    };

    const applications = {
      name: 'kubernetes.applications',
      url: '/applications',
      views: {
        'content@': {
          component: 'kubernetesApplicationsView',
        },
      },
      data: {
        docs: '/user/kubernetes/applications',
      },
    };

    const applicationCreation = {
      name: 'kubernetes.applications.new',
      url: '/new',
      views: {
        'content@': {
          component: 'kubernetesCreateApplicationView',
        },
      },
      onEnter: /* @ngInject */ function endpoint($async, $state, $transition$, Notifications) {
        return $async(async () => {
          try {
            const endpointId = +$transition$.params().endpointId;
            const deploymentOptions = await getDeploymentOptions(endpointId);
            if (deploymentOptions.hideAddWithForm) {
              $state.go('kubernetes.applications', { endpointId });
            }
          } catch (err) {
            Notifications.error('Failed to get deployment options', err);
          }
        });
      },
    };

    const application = {
      name: 'kubernetes.applications.application',
      url: '/:namespace/:name?resource-type&tab',
      views: {
        'content@': {
          component: 'applicationDetailsView',
        },
      },
    };

    const applicationEdit = {
      name: 'kubernetes.applications.application.edit',
      url: '/edit',
      views: {
        'content@': {
          component: 'kubernetesCreateApplicationView',
        },
      },
    };

    const applicationConsole = {
      name: 'kubernetes.applications.application.console',
      url: '/:pod/:container/console',
      views: {
        'content@': {
          component: 'kubernetesConsoleView',
        },
      },
    };

    const applicationLogs = {
      name: 'kubernetes.applications.application.logs',
      url: '/:pod/:container/logs',
      views: {
        'content@': {
          component: 'kubernetesApplicationLogsView',
        },
      },
    };

    const applicationStats = {
      name: 'kubernetes.applications.application.stats',
      url: '/:pod/:container/stats',
      views: {
        'content@': {
          component: 'kubernetesApplicationStatsView',
        },
      },
    };

    const stacks = {
      name: 'kubernetes.stacks',
      url: '/stacks',
      abstract: true,
    };

    const stack = {
      name: 'kubernetes.stacks.stack',
      url: '/:namespace/:name',
      abstract: true,
    };

    const stackLogs = {
      name: 'kubernetes.stacks.stack.logs',
      url: '/logs',
      views: {
        'content@': {
          component: 'kubernetesStackLogsViewAngular',
        },
      },
    };

    const configurations = {
      name: 'kubernetes.configurations',
      url: '/configurations?tab',
      views: {
        'content@': {
          component: 'kubernetesConfigMapsAndSecretsView',
        },
      },
      params: {
        tab: null,
      },
      data: {
        docs: '/user/kubernetes/configurations',
      },
    };

    const configmaps = {
      name: 'kubernetes.configmaps',
      url: '/configmaps',
      abstract: true,
      data: {
        docs: '/user/kubernetes/configurations',
      },
    };

    const configMapCreation = {
      name: 'kubernetes.configmaps.new',
      url: '/new',
      views: {
        'content@': {
          component: 'kubernetesCreateConfigMapView',
        },
      },
      onEnter: /* @ngInject */ function endpoint($async, $state, $transition$, Notifications) {
        return $async(async () => {
          try {
            const endpointId = +$transition$.params().endpointId;
            const deploymentOptions = await getDeploymentOptions(endpointId);
            if (deploymentOptions.hideAddWithForm) {
              $state.go('kubernetes.configurations', { endpointId, tab: 'configmaps' });
            }
          } catch (err) {
            Notifications.error('Failed to get deployment options', err);
          }
        });
      },
    };

    const configMap = {
      name: 'kubernetes.configmaps.configmap',
      url: '/:namespace/:name',
      views: {
        'content@': {
          component: 'kubernetesConfigMapView',
        },
      },
    };

    const secrets = {
      name: 'kubernetes.secrets',
      url: '/secrets',
      abstract: true,
      data: {
        docs: '/user/kubernetes/configurations',
      },
    };

    const secretCreation = {
      name: 'kubernetes.secrets.new',
      url: '/new',
      views: {
        'content@': {
          component: 'kubernetesCreateSecretView',
        },
      },
      onEnter: /* @ngInject */ function endpoint($async, $state, $transition$, Notifications) {
        return $async(async () => {
          try {
            const endpointId = +$transition$.params().endpointId;
            const deploymentOptions = await getDeploymentOptions(endpointId);
            if (deploymentOptions.hideAddWithForm) {
              $state.go('kubernetes.configurations', { endpointId, tab: 'secrets' });
            }
          } catch (err) {
            Notifications.error('Failed to get deployment options', err);
          }
        });
      },
    };

    const secret = {
      name: 'kubernetes.secrets.secret',
      url: '/:namespace/:name',
      views: {
        'content@': {
          component: 'kubernetesSecretView',
        },
      },
    };

    const cluster = {
      name: 'kubernetes.cluster',
      url: '/cluster',
      views: {
        'content@': {
          component: 'kubernetesClusterView',
        },
      },
      data: {
        docs: '/user/kubernetes/cluster',
      },
    };

    const nodes = {
      name: 'kubernetes.cluster.nodes',
      url: '/nodes',
      abstract: true,
    };

    const nodeCreate = {
      name: 'kubernetes.cluster.nodes.new',
      url: '/new',
      views: {
        'content@': {
          component: 'kubernetesNodeCreateView',
        },
      },
    };

    const microk8sNodeStatus = {
      name: 'kubernetes.cluster.node.microk8s-status',
      url: '/microk8s-status',
      views: {
        'content@': {
          component: 'microk8sNodeStatusView',
        },
      },
    };

    const node = {
      name: 'kubernetes.cluster.node',
      url: '/:nodeName',
      views: {
        'content@': {
          component: 'kubernetesNodeView',
        },
      },
    };

    const nodeStats = {
      name: 'kubernetes.cluster.node.stats',
      url: '/stats',
      views: {
        'content@': {
          component: 'kubernetesNodeStatsView',
        },
      },
    };

    const nodeShell = {
      name: 'kubernetes.nodeshell',
      url: '/node-shell?nodeIP',
      views: {
        'content@': {
          component: 'microk8sNodeShellView',
        },
        'sidebar@': {},
      },
      params: {
        nodeIP: null,
      },
    };

    const dashboard = {
      name: 'kubernetes.dashboard',
      url: '/dashboard',
      views: {
        'content@': {
          component: 'kubernetesDashboardView',
        },
      },
      data: {
        docs: '/user/kubernetes/dashboard',
      },
    };

    const deploy = {
      name: 'kubernetes.deploy',
      url: '/deploy?templateId&referrer&tab&buildMethod&chartName',
      views: {
        'content@': {
          component: 'kubernetesDeployView',
        },
      },
      params: {
        yaml: '',
      },
    };

    const resourcePools = {
      name: 'kubernetes.resourcePools',
      url: '/pools',
      views: {
        'content@': {
          component: 'kubernetesResourcePoolsView',
        },
      },
      data: {
        docs: '/user/kubernetes/namespaces',
      },
    };

    const namespaceCreation = {
      name: 'kubernetes.resourcePools.new',
      url: '/new',
      views: {
        'content@': {
          component: 'kubernetesCreateNamespaceView',
        },
      },
    };

    const resourcePool = {
      name: 'kubernetes.resourcePools.resourcePool',
      url: '/:id',
      views: {
        'content@': {
          component: 'kubernetesResourcePoolView',
        },
      },
    };

    const resourcePoolAccess = {
      name: 'kubernetes.resourcePools.resourcePool.access',
      url: '/access',
      views: {
        'content@': {
          component: 'kubernetesResourcePoolAccessView',
        },
      },
    };

    const volumes = {
      name: 'kubernetes.volumes',
      url: '/volumes',
      views: {
        'content@': {
          component: 'kubernetesVolumesView',
        },
      },
      data: {
        docs: '/user/kubernetes/volumes',
      },
    };

    const volume = {
      name: 'kubernetes.volumes.volume',
      url: '/:namespace/:name',
      views: {
        'content@': {
          component: 'kubernetesVolumeView',
        },
      },
    };

    const registries = {
      name: 'kubernetes.registries',
      url: '/registries',
      views: {
        'content@': {
          component: 'endpointRegistriesView',
        },
      },
      data: {
        docs: '/user/kubernetes/cluster/registries',
      },
    };

    const registriesAccess = {
      name: 'kubernetes.registries.access',
      url: '/:id/access',
      views: {
        'content@': {
          component: 'kubernetesRegistryAccessView',
        },
      },
    };

    const endpointKubernetesConfiguration = {
      name: 'kubernetes.cluster.setup',
      url: '/configure',
      views: {
        'content@': {
          component: 'kubernetesConfigureView',
        },
      },
      data: {
        docs: '/user/kubernetes/cluster/setup',
      },
    };

    const endpointKubernetesSecurityConstraint = {
      name: 'kubernetes.cluster.securityConstraint',
      url: '/securityConstraint',
      views: {
        'content@': {
          component: 'kubernetesSecurityConstraintController',
        },
      },
      data: {
        docs: '/user/kubernetes/cluster/security',
      },
    };

    const moreResources = {
      name: 'kubernetes.moreResources',
      url: '/moreResources',
      abstract: true,
    };

    const serviceAccounts = {
      name: 'kubernetes.moreResources.serviceAccounts',
      url: '/serviceAccounts',
      views: {
        'content@': {
          component: 'serviceAccountsView',
        },
      },
      data: {
        docs: '/user/kubernetes/more-resources/service-accounts',
      },
    };

    const clusterRoles = {
      name: 'kubernetes.moreResources.clusterRoles',
      url: '/clusterRoles?tab',
      views: {
        'content@': {
          component: 'clusterRolesView',
        },
      },
      data: {
        docs: '/user/kubernetes/more-resources/cluster-roles',
      },
    };

    const roles = {
      name: 'kubernetes.moreResources.roles',
      url: '/roles?tab',
      views: {
        'content@': {
          component: 'k8sRolesView',
        },
      },
      data: {
        docs: '/user/kubernetes/more-resources/namespace-roles',
      },
    };

    $stateRegistryProvider.register(kubernetes);
    $stateRegistryProvider.register(helmApplication);
    $stateRegistryProvider.register(applications);
    $stateRegistryProvider.register(applicationCreation);
    $stateRegistryProvider.register(application);
    $stateRegistryProvider.register(applicationEdit);
    $stateRegistryProvider.register(applicationConsole);
    $stateRegistryProvider.register(applicationLogs);
    $stateRegistryProvider.register(applicationStats);
    $stateRegistryProvider.register(stacks);
    $stateRegistryProvider.register(stack);
    $stateRegistryProvider.register(stackLogs);
    $stateRegistryProvider.register(configurations);
    $stateRegistryProvider.register(configmaps);
    $stateRegistryProvider.register(configMapCreation);
    $stateRegistryProvider.register(secrets);
    $stateRegistryProvider.register(secretCreation);
    $stateRegistryProvider.register(configMap);
    $stateRegistryProvider.register(secret);
    $stateRegistryProvider.register(cluster);
    $stateRegistryProvider.register(dashboard);
    $stateRegistryProvider.register(deploy);
    $stateRegistryProvider.register(nodes);
    $stateRegistryProvider.register(nodeCreate);
    $stateRegistryProvider.register(node);
    $stateRegistryProvider.register(nodeStats);
    $stateRegistryProvider.register(nodeShell);
    $stateRegistryProvider.register(microk8sNodeStatus);
    $stateRegistryProvider.register(resourcePools);
    $stateRegistryProvider.register(namespaceCreation);
    $stateRegistryProvider.register(resourcePool);
    $stateRegistryProvider.register(resourcePoolAccess);
    $stateRegistryProvider.register(volumes);
    $stateRegistryProvider.register(volume);
    $stateRegistryProvider.register(registries);
    $stateRegistryProvider.register(registriesAccess);
    $stateRegistryProvider.register(endpointKubernetesConfiguration);
    $stateRegistryProvider.register(endpointKubernetesSecurityConstraint);

    $stateRegistryProvider.register(services);
    $stateRegistryProvider.register(ingresses);
    $stateRegistryProvider.register(ingressesCreate);
    $stateRegistryProvider.register(ingressesEdit);

    $stateRegistryProvider.register(moreResources);
    $stateRegistryProvider.register(serviceAccounts);
    $stateRegistryProvider.register(clusterRoles);
    $stateRegistryProvider.register(roles);
  },
]);
