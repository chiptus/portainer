import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';

angular.module('portainer.rbac').directive('authorization', [
  'Authentication',
  '$async',
  'EndpointProvider',
  function (Authentication, $async, EndpointProvider) {
    async function linkAsync(scope, elem, attrs) {
      elem.hide();

      var authorizations = attrs.authorization.split(',');
      for (var i = 0; i < authorizations.length; i++) {
        authorizations[i] = authorizations[i].trim();
      }

      // Check authorizations based on the deployment option settings
      const { hideDeploymentOption } = attrs;
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

      var hasAuthorizations = Authentication.hasAuthorizations(authorizations);
      if (hasAuthorizations) {
        elem.show();
      } else if (!hasAuthorizations && elem[0].tagName === 'A') {
        elem.show();
        elem.addClass('portainer-disabled-link');
      }
    }

    return {
      restrict: 'A',
      link: function (scope, elem, attrs) {
        return $async(linkAsync, scope, elem, attrs);
      },
    };
  },
]);
