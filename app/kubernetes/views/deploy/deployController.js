import angular from 'angular';
import _ from 'lodash-es';
import stripAnsi from 'strip-ansi';

import PortainerError from '@/portainer/error';
import { KubernetesDeployManifestTypes, KubernetesDeployBuildMethods, KubernetesDeployRequestMethods, RepositoryMechanismTypes } from 'Kubernetes/models/deploy';
import { isTemplateVariablesEnabled, renderTemplate } from '@/react/portainer/custom-templates/components/utils';
import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';
import { kubernetes } from '@@/BoxSelector/common-options/deployment-methods';
import { editor, git, customTemplate, url, helm } from '@@/BoxSelector/common-options/build-methods';
import { confirmWebEditorDiscard } from '@@/modals/confirm';
import { parseAutoUpdateResponse, transformAutoUpdateViewModel } from '@/react/portainer/gitops/AutoUpdateFieldset/utils';
import { baseStackWebhookUrl, createWebhookId } from '@/portainer/helpers/webhookHelper';
import { getVariablesFieldDefaultValues } from '@/react/portainer/custom-templates/components/CustomTemplatesVariablesField';

class KubernetesDeployController {
  /* @ngInject */
  constructor($async, $state, $window, Authentication, CustomTemplateService, Notifications, KubernetesResourcePoolService, StackService, UserService) {
    this.$async = $async;
    this.$state = $state;
    this.$window = $window;
    this.Authentication = Authentication;
    this.CustomTemplateService = CustomTemplateService;
    this.Notifications = Notifications;
    this.KubernetesResourcePoolService = KubernetesResourcePoolService;
    this.StackService = StackService;
    this.UserService = UserService;

    this.isTemplateVariablesEnabled = isTemplateVariablesEnabled;

    this.deployOptions = [{ ...kubernetes, value: KubernetesDeployManifestTypes.KUBERNETES }];

    this.methodOptions = [];

    this.currentUser = {
      isAdmin: false,
      id: null,
    };

    let buildMethod = Number(this.$state.params.buildMethod) || KubernetesDeployBuildMethods.GIT;
    if (buildMethod > Object.keys(KubernetesDeployBuildMethods).length) {
      buildMethod = KubernetesDeployBuildMethods.GIT;
    }
    this.state = {
      DeployType: KubernetesDeployManifestTypes.KUBERNETES,
      BuildMethod: buildMethod,
      tabLogsDisabled: true,
      activeTab: 0,
      viewReady: false,
      isEditorDirty: false,
      templateId: null,
      template: null,
      baseWebhookUrl: baseStackWebhookUrl(),
      webhookId: createWebhookId(),
      templateLoadFailed: false,
      isEditorReadOnly: false,
    };

    this.formValues = {
      StackName: '',
      RepositoryURL: '',
      RepositoryURLValid: false,
      RepositoryReferenceName: 'refs/heads/main',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      SaveCredential: true,
      RepositoryGitCredentialID: 0,
      NewCredentialName: '',
      AdditionalFiles: [],
      ComposeFilePathInRepository: '',
      Variables: [],
      AutoUpdate: parseAutoUpdateResponse(),
      TLSSkipVerify: false,
      Name: '', // For helm chart
    };

    this.ManifestDeployTypes = KubernetesDeployManifestTypes;
    this.BuildMethods = KubernetesDeployBuildMethods;

    this.onSelectHelmChart = this.onSelectHelmChart.bind(this);
    this.onChangeTemplateId = this.onChangeTemplateId.bind(this);
    this.deployAsync = this.deployAsync.bind(this);
    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.getNamespacesAsync = this.getNamespacesAsync.bind(this);
    this.onChangeFormValues = this.onChangeFormValues.bind(this);
    this.buildAnalyticsProperties = this.buildAnalyticsProperties.bind(this);
    this.onChangeMethod = this.onChangeMethod.bind(this);
    this.onChangeDeployType = this.onChangeDeployType.bind(this);
    this.onChangeTemplateVariables = this.onChangeTemplateVariables.bind(this);
    this.setStackName = this.setStackName.bind(this);

    const manifestData = this.$state.params.yaml || '';
    if (manifestData) {
      this.formValues.EditorContent = manifestData;
      this.state.BuildMethod = KubernetesDeployBuildMethods.WEB_EDITOR;
    }
  }

  onSelectHelmChart(chart) {
    this.state.selectedHelmChart = chart;
  }

  onChangeTemplateVariables(value) {
    this.onChangeFormValues({ Variables: value });

    this.renderTemplate();
  }

  setStackName(name) {
    this.formValues.StackName = name;
  }

  renderTemplate() {
    if (!this.isTemplateVariablesEnabled) {
      return;
    }

    const rendered = renderTemplate(this.state.templateContent, this.formValues.Variables, this.state.template.Variables);
    this.onChangeFormValues({ EditorContent: rendered });
  }

  buildAnalyticsProperties() {
    const metadata = {
      type: buildLabel(this.state.BuildMethod),
      format: formatLabel(this.state.DeployType),
      role: roleLabel(this.currentUser.isAdmin, this.Authentication.hasAuthorizations(['EndpointResourcesAccess'])),
      'automatic-updates': automaticUpdatesLabel(this.formValues.RepositoryAutomaticUpdates, this.formValues.RepositoryMechanism),
    };

    if (this.state.BuildMethod === KubernetesDeployBuildMethods.GIT) {
      metadata.auth = this.formValues.RepositoryAuthentication;
    }

    return { metadata };

    function automaticUpdatesLabel(repositoryAutomaticUpdates, repositoryMechanism) {
      switch (repositoryAutomaticUpdates && repositoryMechanism) {
        case RepositoryMechanismTypes.INTERVAL:
          return 'polling';
        case RepositoryMechanismTypes.WEBHOOK:
          return 'webhook';
        default:
          return 'off';
      }
    }

    function roleLabel(isAdmin, isEndpointAdmin) {
      if (isAdmin) {
        return 'admin';
      }

      if (isEndpointAdmin) {
        return 'endpoint-admin';
      }

      return 'standard';
    }

    function buildLabel(buildMethod) {
      switch (buildMethod) {
        case KubernetesDeployBuildMethods.GIT:
          return 'git';
        case KubernetesDeployBuildMethods.WEB_EDITOR:
          return 'web-editor';
      }
    }

    function formatLabel(format) {
      switch (format) {
        case KubernetesDeployManifestTypes.COMPOSE:
          return 'compose';
        case KubernetesDeployManifestTypes.KUBERNETES:
          return 'manifest';
      }
    }
  }

  onChangeMethod(method) {
    return this.$async(async () => {
      this.state.BuildMethod = method;
    });
  }

  onChangeDeployType(type) {
    return this.$async(async () => {
      this.state.DeployType = type;
    });
  }

  disableDeploy() {
    const isWebEditorInvalid = this.state.BuildMethod === KubernetesDeployBuildMethods.WEB_EDITOR && _.isEmpty(this.formValues.EditorContent);
    const isURLFormInvalid = this.state.BuildMethod === KubernetesDeployBuildMethods.URL && _.isEmpty(this.formValues.ManifestURL);
    const isCustomTemplateInvalid = this.state.BuildMethod === KubernetesDeployBuildMethods.CUSTOM_TEMPLATE && _.isEmpty(this.formValues.EditorContent);

    const isNamespaceInvalid = _.isEmpty(this.formValues.Namespace);

    return isWebEditorInvalid || isURLFormInvalid || isCustomTemplateInvalid || this.state.actionInProgress || isNamespaceInvalid;
  }

  onChangeFormValues(newValues) {
    return this.$async(async () => {
      this.formValues = {
        ...this.formValues,
        ...newValues,
      };
    });
  }

  onChangeTemplateId(templateId, template) {
    return this.$async(async () => {
      if (!template || (this.state.templateId === templateId && this.state.template === template)) {
        return;
      }

      this.state.templateId = templateId;
      this.state.template = template;

      try {
        try {
          this.state.templateContent = await this.CustomTemplateService.customTemplateFile(templateId, template.GitConfig !== null);
          this.onChangeFileContent(this.state.templateContent);

          this.state.isEditorReadOnly = true;
        } catch (err) {
          this.state.templateLoadFailed = true;
          throw err;
        }

        if (template.Variables && template.Variables.length > 0) {
          const variables = getVariablesFieldDefaultValues(template.Variables);
          this.onChangeTemplateVariables(variables);
        }
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to load template file');
      }
    });
  }

  onChangeFileContent(value) {
    this.formValues.EditorContent = value;
    this.state.isEditorDirty = true;
  }

  displayErrorLog(log) {
    this.errorLog = stripAnsi(log);
    this.state.tabLogsDisabled = false;
    this.state.activeTab = 1;
  }

  async deployAsync() {
    this.errorLog = '';
    this.state.actionInProgress = true;
    const that = this;
    try {
      let method;
      let composeFormat = this.state.DeployType === this.ManifestDeployTypes.COMPOSE;

      switch (this.state.BuildMethod) {
        case this.BuildMethods.GIT:
          method = KubernetesDeployRequestMethods.REPOSITORY;
          break;
        case this.BuildMethods.WEB_EDITOR:
          method = KubernetesDeployRequestMethods.STRING;
          break;
        case KubernetesDeployBuildMethods.CUSTOM_TEMPLATE:
          method = KubernetesDeployRequestMethods.STRING;
          composeFormat = false;
          break;
        case this.BuildMethods.URL:
          method = KubernetesDeployRequestMethods.URL;
          break;
        default:
          throw new PortainerError('Unable to determine build method');
      }

      let deployNamespace = '';

      if (this.formValues.namespace_toggle) {
        deployNamespace = '';
      } else {
        deployNamespace = this.formValues.Namespace;
      }
      const payload = {
        ComposeFormat: composeFormat,
        Namespace: deployNamespace,
        StackName: this.formValues.StackName,
      };

      if (method === KubernetesDeployRequestMethods.REPOSITORY) {
        payload.TLSSkipVerify = this.formValues.TLSSkipVerify;
        payload.RepositoryURL = this.formValues.RepositoryURL;
        payload.RepositoryReferenceName = this.formValues.RepositoryReferenceName;
        payload.RepositoryAuthentication = this.formValues.RepositoryAuthentication ? true : false;
        if (payload.RepositoryAuthentication) {
          // save git credential
          if (this.formValues.SaveCredential && this.formValues.NewCredentialName) {
            await this.UserService.saveGitCredential(
              this.currentUser.id,
              this.formValues.NewCredentialName,
              this.formValues.RepositoryUsername,
              this.formValues.RepositoryPassword
            ).then(function success(data) {
              that.formValues.RepositoryGitCredentialID = data.gitCredential.id;
            });
          }
          payload.RepositoryGitCredentialID = this.formValues.RepositoryGitCredentialID;
          payload.RepositoryUsername = payload.RepositoryGitCredentialID === 0 ? this.formValues.RepositoryUsername : '';
          payload.RepositoryPassword = payload.RepositoryGitCredentialID === 0 ? this.formValues.RepositoryPassword : '';
        }
        payload.ManifestFile = this.formValues.ComposeFilePathInRepository;
        payload.AdditionalFiles = this.formValues.AdditionalFiles;
        payload.AutoUpdate = transformAutoUpdateViewModel(this.formValues.AutoUpdate, this.state.webhookId);
      } else if (method === KubernetesDeployRequestMethods.STRING) {
        payload.StackFileContent = this.formValues.EditorContent;
      } else {
        payload.ManifestURL = this.formValues.ManifestURL;
      }

      await this.StackService.kubernetesDeploy(this.endpoint.Id, method, payload);

      this.Notifications.success('Success', 'Request to deploy manifest successfully submitted');
      this.state.isEditorDirty = false;

      if (this.$state.params.referrer && this.$state.params.tab) {
        this.$state.go(this.$state.params.referrer, { tab: this.$state.params.tab });
        return;
      }

      if (this.$state.params.referrer) {
        this.$state.go(this.$state.params.referrer);
        return;
      }

      this.$state.go('kubernetes.applications');
    } catch (err) {
      this.Notifications.error('Unable to deploy manifest', err, 'Unable to deploy resources');
      if (this.formValues.SaveCredential && this.formValues.NewCredentialName && this.formValues.RepositoryGitCredentialID) {
        this.UserService.deleteGitCredential(this.currentUser.id, this.formValues.RepositoryGitCredentialID);
      }
      this.displayErrorLog(err.err.data.details);
    } finally {
      this.state.actionInProgress = false;
    }
  }

  deploy() {
    return this.$async(this.deployAsync);
  }

  async getNamespacesAsync() {
    try {
      const pools = await this.KubernetesResourcePoolService.get();
      let namespaces = pools.filter((pool) => pool.Namespace.Status === 'Active');
      namespaces = _.map(namespaces, 'Namespace').sort((a, b) => {
        if (a.Name === 'default') {
          return -1;
        }
        if (b.Name === 'default') {
          return 1;
        }
        return 0;
      });
      this.namespaces = namespaces;
      if (this.namespaces.length > 0) {
        this.formValues.Namespace = this.namespaces[0].Name;
      }
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to load namespaces data');
    }
  }

  getNamespaces() {
    return this.$async(this.getNamespacesAsync);
  }

  async uiCanExit() {
    if (this.formValues.EditorContent && this.state.isEditorDirty) {
      return confirmWebEditorDiscard();
    }
  }

  $onInit() {
    return this.$async(async () => {
      this.currentUser.isAdmin = this.Authentication.isAdmin();
      this.currentUser.isPureAdmin = this.Authentication.isPureAdmin();
      const user = this.Authentication.getUserDetails();
      this.currentUser.id = user.ID;

      this.formValues.namespace_toggle = false;
      await this.getNamespaces();

      try {
        this.deploymentOptions = await getDeploymentOptions(this.endpoint.Id);
        if (!this.deploymentOptions.hideWebEditor) {
          this.methodOptions.push({ ...editor, value: KubernetesDeployBuildMethods.WEB_EDITOR });
        }
        if (!this.deploymentOptions.hideFileUpload) {
          this.methodOptions.splice(1, 0, { ...url, value: KubernetesDeployBuildMethods.URL });
        }

        this.methodOptions.push(
          { ...git, value: KubernetesDeployBuildMethods.GIT },
          { ...customTemplate, description: 'Use custom template', value: KubernetesDeployBuildMethods.CUSTOM_TEMPLATE }
        );

        this.Authentication.hasAuthorizations(['HelmInstallChart']) && this.methodOptions.push({ ...helm, value: KubernetesDeployBuildMethods.HELM });

        // the selected method must be available
        if (!this.methodOptions.map((option) => option.value).includes(this.state.Method)) {
          this.state.Method = this.methodOptions[0].value;
        }
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to get deployment options');
      }

      if (this.$state.params.templateId) {
        const templateId = parseInt(this.$state.params.templateId, 10);
        if (templateId && !Number.isNaN(templateId) && !this.deploymentOptions.hideWebEditor) {
          this.state.BuildMethod = KubernetesDeployBuildMethods.CUSTOM_TEMPLATE;
          this.state.templateId = templateId;
        }
      }

      this.state.viewReady = true;

      this.$window.onbeforeunload = () => {
        if (this.formValues.EditorContent && this.state.isEditorDirty) {
          return '';
        }
      };
    });
  }

  $onDestroy() {
    this.state.isEditorDirty = false;
  }
}

export default KubernetesDeployController;
angular.module('portainer.kubernetes').controller('KubernetesDeployController', KubernetesDeployController);
