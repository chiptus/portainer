import angular from 'angular';
import YAML from 'yaml';

import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';

class KubernetesYamlInspectorController {
  /* @ngInject */

  constructor(clipboard, Authentication, EndpointProvider, $async) {
    this.clipboard = clipboard;
    this.Authentication = Authentication;
    this.expanded = false;
    this.EndpointProvider = EndpointProvider;
    this.$async = $async;

    this.yaml = '';

    this.onChange = this.onChange.bind(this);

    this.deploymentOptions = {
      hideWebEditor: true,
    };

    this.loading = true;
  }

  cleanYamlUnwantedFields(yml) {
    try {
      const ymls = yml.split('---');
      const cleanYmls = ymls.map((yml) => {
        const y = YAML.parse(yml);
        if (y.metadata) {
          delete y.metadata.managedFields;
          delete y.metadata.resourceVersion;
        }
        return YAML.stringify(y);
      });
      return cleanYmls.join('---\n');
    } catch (e) {
      return yml;
    }
  }

  copyYAML() {
    this.clipboard.copyText(this.yaml);
    $('#copyNotificationYAML').show().fadeOut(2500);
  }

  toggleYAMLInspectorExpansion() {
    let selector = 'kubernetes-yaml-inspector code-editor > div.CodeMirror';
    let height = this.expanded ? '500px' : '80vh';
    $(selector).css({ height: height });
    this.expanded = !this.expanded;
  }

  onChange(yml) {
    this.yaml = yml;
  }

  $onInit() {
    return this.$async(async () => {
      this.data = this.cleanYamlUnwantedFields(this.data);
      this.yaml = this.data;

      try {
        const endpoint = this.EndpointProvider.currentEndpoint();
        this.deploymentOptions = await getDeploymentOptions(endpoint.Id);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve deployment options');
      }

      this.isAuthorized = this.Authentication.hasAuthorizations(['K8sYAMLW']) && this.authorised && !this.deploymentOptions.hideWebEditor;

      this.loading = false;
    });
  }
}

export default KubernetesYamlInspectorController;
angular.module('portainer.kubernetes').controller('KubernetesYamlInspectorController', KubernetesYamlInspectorController);
