import angular from 'angular';

import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';
import { kubeCustomTemplatesView } from './kube-custom-templates-view';
import { kubeEditCustomTemplateView } from './kube-edit-custom-template-view';

export default angular
  .module('portainer.kubernetes.custom-templates', [])
  .config(config)
  .component('kubeCustomTemplatesView', kubeCustomTemplatesView)
  .component('kubeEditCustomTemplateView', kubeEditCustomTemplateView).name;

function config($stateRegistryProvider) {
  const templates = {
    name: 'kubernetes.templates',
    url: '/templates',
    abstract: true,
  };

  const customTemplates = {
    name: 'kubernetes.templates.custom',
    url: '/custom',

    views: {
      'content@': {
        component: 'kubeCustomTemplatesView',
      },
    },
    onEnter: /* @ngInject */ function endpoint($async, $state, $transition$, Notifications) {
      return $async(async () => {
        try {
          const endpointId = +$transition$.params().endpointId;
          const deploymentOptions = await getDeploymentOptions(endpointId);
          if (deploymentOptions.hideWebEditor) {
            $state.go('kubernetes.dashboard', { endpointId });
          }
        } catch (err) {
          Notifications.error('Failed to get deployment options', err);
        }
      });
    },
    data: {
      docs: '/user/kubernetes/templates',
    },
  };

  const customTemplatesNew = {
    name: 'kubernetes.templates.custom.new',
    url: '/new?fileContent',

    views: {
      'content@': {
        component: 'createCustomTemplatesView',
      },
    },
    params: {
      fileContent: '',
    },
    onEnter: /* @ngInject */ function endpoint($async, $state, $transition$, Notifications) {
      return $async(async () => {
        try {
          const endpointId = +$transition$.params().endpointId;
          const deploymentOptions = await getDeploymentOptions(endpointId);
          if (deploymentOptions.hideWebEditor) {
            $state.go('kubernetes.templates.custom', { endpointId });
          }
        } catch (err) {
          Notifications.error('Failed to get deployment options', err);
        }
      });
    },
  };

  const customTemplatesEdit = {
    name: 'kubernetes.templates.custom.edit',
    url: '/:id',

    views: {
      'content@': {
        component: 'kubeEditCustomTemplateView',
      },
    },
    onEnter: /* @ngInject */ function endpoint($async, $state, $transition$, Notifications) {
      return $async(async () => {
        try {
          const endpointId = +$transition$.params().endpointId;
          const deploymentOptions = await getDeploymentOptions(endpointId);
          if (deploymentOptions.hideWebEditor) {
            $state.go('kubernetes.dashboard', { endpointId });
          }
        } catch (err) {
          Notifications.error('Failed to get deployment options', err);
        }
      });
    },
  };

  $stateRegistryProvider.register(templates);
  $stateRegistryProvider.register(customTemplates);
  $stateRegistryProvider.register(customTemplatesNew);
  $stateRegistryProvider.register(customTemplatesEdit);
}
