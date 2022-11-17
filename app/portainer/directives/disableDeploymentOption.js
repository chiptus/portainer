import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';
import { getGlobalDeploymentOptions } from '@/react/portainer/settings/settings.service';

angular.module('portainer.app').directive('disableDeploymentOption', [
  '$async',
  'EndpointProvider',
  function ($async, EndpointProvider) {
    async function linkAsync(scope, elem, attrs) {
      const { disableDeploymentOption, disableDeploymentScope } = attrs;

      // if there's a hide-deployment-option attribute, check if the alement should be hidden
      if (disableDeploymentOption) {
        const endpoint = EndpointProvider.currentEndpoint();
        let deploymentOptions;
        // decide whether to use local (environment) settings or global settings
        if (disableDeploymentScope === 'global') {
          deploymentOptions = await getGlobalDeploymentOptions();
        } else {
          deploymentOptions = await getDeploymentOptions(endpoint.Id);
        }
        // hide the element if the deployment option is in the attribute and is hidden in settings
        if (deploymentOptions) {
          if (disableDeploymentOption == 'form' && !deploymentOptions.hideAddWithForm) {
            return;
          }
          if (disableDeploymentOption == 'webEditor' && !deploymentOptions.hideWebEditor) {
            return;
          }
          if (disableDeploymentOption == 'fileUpload' && !deploymentOptions.hideFileUpload) {
            return;
          }
        }
      }

      attrs.$set('read-only', true);
      attrs.$set('disabled', true);
    }

    return {
      restrict: 'A',
      link: function (scope, elem, attrs) {
        return $async(linkAsync, scope, elem, attrs);
      },
    };
  },
]);
