import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';

angular.module('portainer.app').directive('hideDeploymentOption', [
  '$async',
  'EndpointProvider',
  function ($async, EndpointProvider) {
    async function linkAsync(scope, elem, attrs) {
      // handle full auth logic in the authorization directive (deployment option logic is there too)
      const { authorization, hideDeploymentOption } = attrs;
      if (authorization) {
        return;
      }

      // otherwise handle the hide-deployment-option logic here
      elem.hide();

      // if there's a hide-deployment-option attribute, check if the alement should be hidden
      if (hideDeploymentOption) {
        const endpoint = EndpointProvider.currentEndpoint();
        const deploymentOptions = await getDeploymentOptions(endpoint.Id);
        // hide the element if the deployment option is in the attribute and is hidden in settings
        if (deploymentOptions) {
          if (hideDeploymentOption == 'form' && deploymentOptions.hideAddWithForm) {
            return;
          }
          if (hideDeploymentOption == 'webEditor' && deploymentOptions.hideWebEditor) {
            return;
          }
          if (hideDeploymentOption == 'fileUpload' && deploymentOptions.hideFileUpload) {
            return;
          }
        }
      }

      elem.show();
    }

    return {
      restrict: 'A',
      link: function (scope, elem, attrs) {
        return $async(linkAsync, scope, elem, attrs);
      },
    };
  },
]);
