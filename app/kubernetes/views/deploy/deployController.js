import angular from 'angular';
import uuidv4 from 'uuid/v4';
import _ from 'lodash-es';
import stripAnsi from 'strip-ansi';

import PortainerError from '@/portainer/error';
import { KubernetesDeployManifestTypes, KubernetesDeployBuildMethods, KubernetesDeployRequestMethods, RepositoryMechanismTypes } from 'Kubernetes/models/deploy';
import { renderTemplate } from '@/react/portainer/custom-templates/components/utils';
import { isBE } from '@/portainer/feature-flags/feature-flags.service';
import { compose, kubernetes } from '@@/BoxSelector/common-options/deployment-methods';
import { editor, git, template, url } from '@@/BoxSelector/common-options/build-methods';

class KubernetesDeployController {
  /* @ngInject */
  constructor(
    $async,
    $state,
    $window,
    Authentication,
    CustomTemplateService,
    ModalService,
    Notifications,
    KubernetesResourcePoolService,
    StackService,
    WebhookHelper,
    UserService
  ) {
    this.$async = $async;
    this.$state = $state;
    this.$window = $window;
    this.Authentication = Authentication;
    this.CustomTemplateService = CustomTemplateService;
    this.ModalService = ModalService;
    this.Notifications = Notifications;
    this.KubernetesResourcePoolService = KubernetesResourcePoolService;
    this.StackService = StackService;
    this.WebhookHelper = WebhookHelper;
    this.UserService = UserService;
    this.DeployMethod = 'manifest';

    this.isTemplateVariablesEnabled = isBE;

    this.deployOptions = [
      { ...kubernetes, value: KubernetesDeployManifestTypes.KUBERNETES },
      { ...compose, value: KubernetesDeployManifestTypes.COMPOSE },
    ];

    this.methodOptions = [
      { ...git, value: KubernetesDeployBuildMethods.GIT },
      { ...editor, value: KubernetesDeployBuildMethods.WEB_EDITOR },
      { ...url, value: KubernetesDeployBuildMethods.URL },
      { ...template, description: 'Use custom template', value: KubernetesDeployBuildMethods.CUSTOM_TEMPLATE },
    ];

    this.state = {
      DeployType: KubernetesDeployManifestTypes.KUBERNETES,
      BuildMethod: KubernetesDeployBuildMethods.GIT,
      tabLogsDisabled: true,
      activeTab: 0,
      viewReady: false,
      isEditorDirty: false,
      templateId: null,
      template: null,
    };

    this.formValues = {
      StackName: '',
      RepositoryURL: '',
      RepositoryURLValid: false,
      RepositoryReferenceName: 'refs/heads/main',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      SelectedGitCredential: null,
      GitCredentials: [],
      SaveCredential: true,
      RepositoryGitCredentialID: 0,
      NewCredentialName: '',
      NewCredentialNameExist: false,
      NewCredentialNameInvalid: false,
      AdditionalFiles: [],
      ComposeFilePathInRepository: '',
      RepositoryAutomaticUpdates: false,
      RepositoryAutomaticUpdatesForce: false,
      RepositoryMechanism: RepositoryMechanismTypes.INTERVAL,
      RepositoryFetchInterval: '5m',
      RepositoryWebhookURL: WebhookHelper.returnStackWebhookUrl(uuidv4()),
      Variables: {},
    };

    this.ManifestDeployTypes = KubernetesDeployManifestTypes;
    this.BuildMethods = KubernetesDeployBuildMethods;

    this.onChangeTemplateId = this.onChangeTemplateId.bind(this);
    this.deployAsync = this.deployAsync.bind(this);
    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.getNamespacesAsync = this.getNamespacesAsync.bind(this);
    this.onChangeFormValues = this.onChangeFormValues.bind(this);
    this.buildAnalyticsProperties = this.buildAnalyticsProperties.bind(this);
    this.onChangeMethod = this.onChangeMethod.bind(this);
    this.onChangeDeployType = this.onChangeDeployType.bind(this);
    this.onChangeTemplateVariables = this.onChangeTemplateVariables.bind(this);
    this.onChangeGitCredential = this.onChangeGitCredential.bind(this);
  }

  onChangeTemplateVariables(value) {
    this.onChangeFormValues({ Variables: value });

    this.renderTemplate();
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
      role: roleLabel(this.Authentication.isAdmin(), this.Authentication.hasAuthorizations(['EndpointResourcesAccess'])),
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
    this.state.BuildMethod = method;
  }

  onChangeDeployType(type) {
    this.state.DeployType = type;
    if (type == this.ManifestDeployTypes.COMPOSE) {
      this.DeployMethod = 'compose';
    } else {
      this.DeployMethod = 'manifest';
    }
  }

  disableDeploy() {
    const isGitFormInvalid =
      this.state.BuildMethod === KubernetesDeployBuildMethods.GIT &&
      (!this.formValues.RepositoryURL ||
        !this.formValues.ComposeFilePathInRepository ||
        (this.formValues.RepositoryAuthentication && !this.formValues.RepositoryPassword && this.formValues.RepositoryGitCredentialID === 0) ||
        (this.formValues.RepositoryAuthentication &&
          this.formValues.RepositoryPassword &&
          this.formValues.SaveCredential &&
          (!this.formValues.NewCredentialName || this.formValues.NewCredentialNameExist)));

    const isWebEditorInvalid =
      this.state.BuildMethod === KubernetesDeployBuildMethods.WEB_EDITOR && _.isEmpty(this.formValues.EditorContent) && _.isEmpty(this.formValues.Namespace);
    const isURLFormInvalid = this.state.BuildMethod == KubernetesDeployBuildMethods.WEB_EDITOR.URL && _.isEmpty(this.formValues.ManifestURL);

    const isNamespaceInvalid = _.isEmpty(this.formValues.Namespace);

    return !this.formValues.StackName || isGitFormInvalid || isWebEditorInvalid || isURLFormInvalid || this.state.actionInProgress || isNamespaceInvalid;
  }

  onChangeFormValues(newValues) {
    return this.$async(async () => {
      this.formValues = {
        ...this.formValues,
        ...newValues,
      };
      const existGitCredential = this.formValues.GitCredentials.find((x) => x.name === this.formValues.NewCredentialName);
      this.formValues.NewCredentialNameExist = existGitCredential ? true : false;
      this.formValues.NewCredentialNameInvalid = this.formValues.NewCredentialName && !this.formValues.NewCredentialName.match(/^[-_a-z0-9]+$/) ? true : false;
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
        const fileContent = await this.CustomTemplateService.customTemplateFile(templateId);
        this.state.templateContent = fileContent;
        this.onChangeFileContent(fileContent);

        if (template.Variables && template.Variables.length > 0) {
          const variables = Object.fromEntries(template.Variables.map((variable) => [variable.name, '']));
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
        const userDetails = this.Authentication.getUserDetails();

        payload.RepositoryURL = this.formValues.RepositoryURL;
        payload.RepositoryReferenceName = this.formValues.RepositoryReferenceName;
        payload.RepositoryAuthentication = this.formValues.RepositoryAuthentication ? true : false;
        if (payload.RepositoryAuthentication) {
          // save git credential
          if (this.formValues.SaveCredential && this.formValues.NewCredentialName) {
            await this.UserService.saveGitCredential(
              userDetails.ID,
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
        if (this.formValues.RepositoryAutomaticUpdates) {
          payload.AutoUpdate = {
            ForceUpdate: this.formValues.RepositoryAutomaticUpdatesForce,
          };
          if (this.formValues.RepositoryMechanism === RepositoryMechanismTypes.INTERVAL) {
            payload.AutoUpdate.Interval = this.formValues.RepositoryFetchInterval;
          } else if (this.formValues.RepositoryMechanism === RepositoryMechanismTypes.WEBHOOK) {
            payload.AutoUpdate.Webhook = this.formValues.RepositoryWebhookURL.split('/').reverse()[0];
          }
        }
      } else if (method === KubernetesDeployRequestMethods.STRING) {
        payload.StackFileContent = this.formValues.EditorContent;
      } else {
        payload.ManifestURL = this.formValues.ManifestURL;
      }

      await this.StackService.kubernetesDeploy(this.endpoint.Id, method, payload);

      this.Notifications.success('Success', 'Manifest successfully deployed');
      this.state.isEditorDirty = false;
      this.$state.go('kubernetes.applications');
    } catch (err) {
      this.Notifications.error('Unable to deploy manifest', err, 'Unable to deploy resources');
      const userDetails = this.Authentication.getUserDetails();
      if (this.formValues.SaveCredential && this.formValues.NewCredentialName && this.formValues.RepositoryGitCredentialID) {
        this.UserService.deleteGitCredential(userDetails.ID, this.formValues.RepositoryGitCredentialID);
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
      const namespaces = _.map(pools, 'Namespace').sort((a, b) => {
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
      return this.ModalService.confirmWebEditorDiscard();
    }
  }

  $onInit() {
    return this.$async(async () => {
      this.formValues.namespace_toggle = false;
      await this.getNamespaces();

      if (this.$state.params.templateId) {
        const templateId = parseInt(this.$state.params.templateId, 10);
        if (templateId && !Number.isNaN(templateId)) {
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

      try {
        this.formValues.GitCredentials = await this.UserService.getGitCredentials(this.Authentication.getUserDetails().ID);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve user saved git credentials');
      }
    });
  }

  $onDestroy() {
    this.state.isEditorDirty = false;
  }

  onChangeGitCredential(selectedGitCredential) {
    return this.$async(async () => {
      if (selectedGitCredential) {
        this.formValues.SelectedGitCredential = selectedGitCredential;
        this.formValues.RepositoryGitCredentialID = Number(selectedGitCredential.id);
        this.formValues.RepositoryUsername = selectedGitCredential.username;
        this.formValues.SaveGitCredential = false;
        this.formValues.NewCredentialName = '';
      } else {
        this.formValues.SelectedGitCredential = null;
        this.formValues.RepositoryUsername = '';
        this.formValues.RepositoryGitCredentialID = 0;
      }

      this.formValues.RepositoryPassword = '';
    });
  }
}

export default KubernetesDeployController;
angular.module('portainer.kubernetes').controller('KubernetesDeployController', KubernetesDeployController);
