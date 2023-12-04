import { DeploymentType, EditorType } from '@/react/edge/edge-stacks/types';
import { getValidEditorTypes } from '@/react/edge/edge-stacks/utils';
import { STACK_NAME_VALIDATION_REGEX } from '@/react/constants';
import { confirmWebEditorDiscard } from '@@/modals/confirm';
import { baseEdgeStackWebhookUrl, createWebhookId } from '@/portainer/helpers/webhookHelper';
import { parseAutoUpdateResponse, transformAutoUpdateViewModel } from '@/react/portainer/gitops/AutoUpdateFieldset/utils';
import { EnvironmentType } from '@/react/portainer/environments/types';
import { getCustomTemplate } from '@/react/portainer/templates/custom-templates/queries/useCustomTemplate';
import { notifyError } from '@/portainer/services/notifications';
import { getCustomTemplateFile } from '@/react/portainer/templates/custom-templates/queries/useCustomTemplateFile';
import { getDefaultStaggerConfig } from '@/react/edge/edge-stacks/components/StaggerFieldset.types';
import { toGitFormModel } from '@/react/portainer/gitops/types';
import { StackType } from '@/react/common/stacks/types';
import { applySetStateAction } from '@/react-tools/apply-set-state-action';
import { getVariablesFieldDefaultValues } from '@/react/portainer/custom-templates/components/CustomTemplatesVariablesField';
import { renderTemplate } from '@/react/portainer/custom-templates/components/utils';
import { parseRegistries } from '@/react/edge/edge-stacks/queries/useParseRegistries';
import { getInitialTemplateValues } from '@/react/edge/edge-stacks/CreateView/TemplateFieldset';

export default class CreateEdgeStackViewController {
  /* @ngInject */
  constructor($state, $window, EdgeStackService, EdgeGroupService, Notifications, $async, $scope, RegistryService, UserService, Authentication) {
    Object.assign(this, { $state, $window, EdgeStackService, EdgeGroupService, Notifications, $async, RegistryService, $scope, UserService, Authentication });

    this.formValues = {
      Name: '',
      StackFileContent: '',
      StackFile: null,
      RepositoryURL: '',
      RepositoryURLValid: false,
      RepositoryReferenceName: 'refs/heads/main',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      SaveCredential: true,
      RepositoryGitCredentialID: 0,
      NewCredentialName: '',
      ComposeFilePathInRepository: '',
      Groups: [],
      DeploymentType: EditorType.Compose,
      RegistryOptions: [],
      PrivateRegistryId: null,
      UseManifestNamespaces: false,
      PrePullImage: false,
      RetryDeploy: false,
      TLSSkipVerify: false,
      AutoUpdate: parseAutoUpdateResponse(),
      webhookEnabled: false,
      SupportRelativePath: false,
      FilesystemPath: '',
      versions: [1],
      envVars: [],
      staggerConfig: getDefaultStaggerConfig(),
    };

    this.EditorType = EditorType;
    this.EnvironmentType = EnvironmentType;

    this.state = {
      Method: 'editor',
      formValidationError: '',
      actionInProgress: false,
      StackType: null,
      isEditorDirty: false,
      endpointTypes: [],
      baseWebhookUrl: baseEdgeStackWebhookUrl(),
      webhookId: createWebhookId(),
      isEdit: false,
      templateValues: getInitialTemplateValues(),
    };

    this.edgeGroups = null;

    $scope.STACK_NAME_VALIDATION_REGEX = STACK_NAME_VALIDATION_REGEX;

    this.createStack = this.createStack.bind(this);
    this.validateForm = this.validateForm.bind(this);
    this.createStackByMethod = this.createStackByMethod.bind(this);
    this.createStackFromFileContent = this.createStackFromFileContent.bind(this);
    this.createStackFromFileUpload = this.createStackFromFileUpload.bind(this);
    this.createStackFromGitRepository = this.createStackFromGitRepository.bind(this);
    this.onChangeGroups = this.onChangeGroups.bind(this);
    this.matchRegistry = this.matchRegistry.bind(this);
    this.clearRegistries = this.clearRegistries.bind(this);
    this.selectedRegistry = this.selectedRegistry.bind(this);
    this.hasType = this.hasType.bind(this);
    this.onChangeDeploymentType = this.onChangeDeploymentType.bind(this);
    this.onChangePrePullImage = this.onChangePrePullImage.bind(this);
    this.onChangeRetryDeploy = this.onChangeRetryDeploy.bind(this);
    this.onChangeWebhookState = this.onChangeWebhookState.bind(this);
    this.onEnvVarChange = this.onEnvVarChange.bind(this);
    this.onChangeStaggerConfig = this.onChangeStaggerConfig.bind(this);
    this.setTemplateValues = this.setTemplateValues.bind(this);
    this.onChangeTemplate = this.onChangeTemplate.bind(this);
  }

  /**
   * @param {import('react').SetStateAction<import('@/react/edge/edge-stacks/CreateView/TemplateFieldset').Values>} templateAction
   */
  setTemplateValues(templateAction) {
    return this.$async(async () => {
      const newTemplateValues = applySetStateAction(templateAction, this.state.templateValues);
      const oldTemplateId = this.state.templateValues.template && this.state.templateValues.template.Id;
      const newTemplateId = newTemplateValues.template && newTemplateValues.template.Id;
      this.state.templateValues = newTemplateValues;
      if (newTemplateId !== oldTemplateId) {
        await this.onChangeTemplate(newTemplateValues.template);
      }

      const newFile = renderTemplate(this.state.templateValues.file, this.state.templateValues.variables, this.state.templateValues.template.Variables);

      this.formValues.StackFileContent = newFile;
    });
  }

  onChangeTemplate(template) {
    return this.$async(async () => {
      if (!template) {
        return;
      }

      this.state.templateValues.template = template;
      this.state.templateValues.variables = getVariablesFieldDefaultValues(template.Variables);

      const fileContent = await getCustomTemplateFile({ id: template.Id, git: !!template.GitConfig });
      this.state.templateValues.file = fileContent;

      this.formValues = {
        ...this.formValues,
        DeploymentType: template.Type === StackType.Kubernetes ? DeploymentType.Kubernetes : DeploymentType.Compose,
        ...toGitFormModel(template.GitConfig),
        ...(template.EdgeSettings
          ? {
              PrePullImage: template.EdgeSettings.PrePullImage || false,
              RetryDeploy: template.EdgeSettings.RetryDeploy || false,
              PrivateRegistryId: template.EdgeSettings.PrivateRegistryId || null,
              staggerConfig: template.EdgeSettings.StaggerConfig || getDefaultStaggerConfig(),
              ...template.EdgeSettings.RelativePathSettings,
            }
          : {}),
      };
    });
  }

  onEnvVarChange(envVars) {
    return this.$scope.$evalAsync(() => {
      this.formValues.envVars = envVars;
    });
  }

  onChangeStaggerConfig(staggerConfig) {
    return this.$scope.$evalAsync(() => {
      this.formValues.staggerConfig = { ...this.formValues.staggerConfig, ...staggerConfig };
    });
  }

  buildAnalyticsProperties() {
    const format = 'compose';
    const metadata = { type: methodLabel(this.state.Method), format };

    if (metadata.type === 'template') {
      metadata.templateName = this.state.selectedTemplate && this.state.selectedTemplate.title;
    }

    return { metadata };

    function methodLabel(method) {
      switch (method) {
        case 'editor':
          return 'web-editor';
        case 'repository':
          return 'git';
        case 'upload':
          return 'file-upload';
        case 'template':
          return 'template';
      }
    }
  }

  async uiCanExit() {
    if (this.state.Method === 'editor' && this.formValues.StackFileContent && this.state.isEditorDirty) {
      return confirmWebEditorDiscard();
    }
  }

  async preSelectTemplate(templateId) {
    return this.$async(async () => {
      try {
        this.state.Method = 'template';
        const template = await getCustomTemplate(templateId);

        this.setTemplateValues({ template });
      } catch (e) {
        notifyError('Failed loading template', e);
      }
    });
  }

  async $onInit() {
    // Initial Registry Options for selector
    this.formValues.RegistryOptions = await this.RegistryService.registries();

    this.errorMessage = '';

    try {
      this.edgeGroups = await this.EdgeGroupService.groups();
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve Edge groups');
    }

    const templateId = this.$state.params.templateId;
    if (templateId) {
      this.preSelectTemplate(templateId);
    }

    this.$window.onbeforeunload = () => {
      if (this.state.Method === 'editor' && this.formValues.StackFileContent && this.state.isEditorDirty) {
        return '';
      }
    };
  }

  onChangeWebhookState(state) {
    this.$scope.$evalAsync(() => {
      this.formValues.webhookEnabled = state;
    });
  }

  $onDestroy() {
    this.state.isEditorDirty = false;
  }

  checkRegistries(registries) {
    return registries.every((value) => value === registries[0]);
  }

  clearRegistries() {
    this.formValues.PrivateRegistryId = null;
  }

  selectedRegistry(selectedRegistry) {
    return this.$async(async () => {
      this.formValues.PrivateRegistryId = selectedRegistry;
    });
  }

  matchRegistry() {
    return this.$async(async () => {
      this.errorMessage = '';
      const method = getMethod(this.state.Method, this.state.templateValues.template);
      if (method === 'repository' || !this.formValues.StackFileContent || !this.formValues.StackFile) {
        return;
      }

      this.state.actionInProgress = true;
      try {
        const registries = await parseRegistries({ fileContent: this.formValues.StackFileContent, file: this.formValues.StackFile });

        if (registries.length === 0) {
          this.formValues.PrivateRegistryId = null;
        }
        const validRegistry = this.checkRegistries(registries);
        if (validRegistry) {
          this.formValues.PrivateRegistryId = registries[0];
        } else {
          this.formValues.PrivateRegistryId = null;
          this.errorMessage = ' Images need to be from a single registry, please edit and reload';
        }
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve registries');
      } finally {
        this.state.actionInProgress = false;
      }
    });
  }

  createStack() {
    return this.$async(async () => {
      const name = this.formValues.Name;
      const method = getMethod(this.state.Method, this.state.templateValues.template);

      if (!this.validateForm(method)) {
        return;
      }

      this.state.actionInProgress = true;
      try {
        await this.createStackByMethod(name, method);

        this.Notifications.success('Success', 'Stack successfully deployed');
        this.state.isEditorDirty = false;
        this.$state.go('edge.stacks');
      } catch (err) {
        this.Notifications.error('Deployment error', err, 'Unable to deploy stack');
        if (this.formValues.SaveCredential && this.formValues.NewCredentialName && this.formValues.RepositoryGitCredentialID) {
          const userDetails = this.Authentication.getUserDetails();
          this.UserService.deleteGitCredential(userDetails.ID, this.formValues.RepositoryGitCredentialID);
        }
      } finally {
        this.state.actionInProgress = false;
      }
    });
  }

  onChangeGroups(groups) {
    return this.$scope.$evalAsync(() => {
      this.formValues.Groups = groups;

      this.checkIfEndpointTypes(groups);
    });
  }

  checkIfEndpointTypes(groups) {
    return this.$scope.$evalAsync(() => {
      const edgeGroups = groups.map((id) => this.edgeGroups.find((e) => e.Id === id));
      this.state.endpointTypes = edgeGroups.flatMap((group) => group.EndpointTypes);
      this.selectValidDeploymentType();
    });
  }

  selectValidDeploymentType() {
    const validTypes = getValidEditorTypes(this.state.endpointTypes);

    if (!validTypes.includes(this.formValues.DeploymentType)) {
      this.onChangeDeploymentType(validTypes[0]);
    }
  }

  hasType(envType) {
    return this.state.endpointTypes.includes(envType);
  }

  validateForm(method) {
    this.state.formValidationError = '';

    if (method === 'editor' && this.formValues.StackFileContent === '') {
      this.state.formValidationError = 'Stack file content must not be empty';
      return;
    }

    return true;
  }

  createStackByMethod(name, method) {
    switch (method) {
      case 'editor':
        return this.createStackFromFileContent(name);
      case 'upload':
        return this.createStackFromFileUpload(name);
      case 'repository':
        return this.createStackFromGitRepository(name);
    }
  }

  createStackFromFileContent(name) {
    const { StackFileContent, Groups, DeploymentType, PrivateRegistryId, UseManifestNamespaces, PrePullImage, RetryDeploy, webhookEnabled, envVars, staggerConfig } =
      this.formValues;

    let webhookId = '';
    if (webhookEnabled) {
      webhookId = this.state.webhookId;
    }

    return this.EdgeStackService.createStackFromFileContent({
      name,
      StackFileContent,
      EdgeGroups: Groups,
      DeploymentType,
      Registries: [PrivateRegistryId],
      UseManifestNamespaces,
      PrePullImage,
      RetryDeploy,
      webhook: webhookId,
      envVars,
      staggerConfig,
    });
  }

  createStackFromFileUpload(name) {
    const { StackFile, Groups, DeploymentType, PrivateRegistryId, UseManifestNamespaces, PrePullImage, RetryDeploy, webhookEnabled, envVars, staggerConfig } = this.formValues;
    let webhookId = '';
    if (webhookEnabled) {
      webhookId = this.state.webhookId;
    }

    return this.EdgeStackService.createStackFromFileUpload(
      {
        Name: name,
        EdgeGroups: Groups,
        DeploymentType,
        Registries: [PrivateRegistryId],
        UseManifestNamespaces,
        PrePullImage,
        RetryDeploy,
        webhook: webhookId,
        envVars,
        staggerConfig,
      },
      StackFile
    );
  }

  async createStackFromGitRepository(name) {
    const {
      Groups,
      DeploymentType,
      PrivateRegistryId,
      UseManifestNamespaces,
      PrePullImage,
      RetryDeploy,
      SupportRelativePath,
      FilesystemPath,
      SupportPerDeviceConfigs,
      PerDeviceConfigsMatchType,
      PerDeviceConfigsGroupMatchType,
      PerDeviceConfigsPath,
      envVars,
      staggerConfig,
    } = this.formValues;

    if (this.formValues.SaveCredential && this.formValues.NewCredentialName) {
      const userDetails = this.Authentication.getUserDetails();
      const that = this;
      await this.UserService.saveGitCredential(userDetails.ID, this.formValues.NewCredentialName, this.formValues.RepositoryUsername, this.formValues.RepositoryPassword).then(
        function success(data) {
          that.formValues.RepositoryGitCredentialID = data.gitCredential.id;
        }
      );
    }

    const autoUpdate = transformAutoUpdateViewModel(this.formValues.AutoUpdate, this.state.webhookId);

    const repositoryOptions = {
      RepositoryURL: this.formValues.RepositoryURL,
      RepositoryReferenceName: this.formValues.RepositoryReferenceName,
      FilePathInRepository: this.formValues.ComposeFilePathInRepository,
      RepositoryAuthentication: this.formValues.RepositoryAuthentication,
      RepositoryUsername: this.formValues.RepositoryUsername,
      RepositoryPassword: this.formValues.RepositoryPassword,
      RepositoryGitCredentialID: this.formValues.RepositoryGitCredentialID,
      TLSSkipVerify: this.formValues.TLSSkipVerify,
    };

    return this.EdgeStackService.createStackFromGitRepository(
      {
        name,
        EdgeGroups: Groups,
        DeploymentType,
        Registries: [PrivateRegistryId],
        UseManifestNamespaces,
        PrePullImage,
        RetryDeploy,
        AutoUpdate: autoUpdate,
        SupportRelativePath: SupportRelativePath || SupportPerDeviceConfigs,
        FilesystemPath,
        SupportPerDeviceConfigs,
        PerDeviceConfigsMatchType,
        PerDeviceConfigsGroupMatchType,
        PerDeviceConfigsPath,
        envVars,
        staggerConfig,
      },
      repositoryOptions
    );
  }

  onChangeDeploymentType(deploymentType) {
    return this.$scope.$evalAsync(() => {
      this.formValues.DeploymentType = deploymentType;
      this.state.Method = 'editor';
      this.formValues.StackFileContent = '';
      this.state.templateValues = getInitialTemplateValues();
    });
  }

  onChangePrePullImage(value) {
    return this.$scope.$evalAsync(() => {
      this.formValues.PrePullImage = value;
    });
  }

  onChangeRetryDeploy(value) {
    return this.$scope.$evalAsync(() => {
      this.formValues.RetryDeploy = value;
    });
  }

  formIsInvalid() {
    return (
      this.form.$invalid ||
      !this.formValues.Groups.length ||
      (['template', 'editor'].includes(this.state.Method) && !this.formValues.StackFileContent) ||
      ('upload' === this.state.Method && !this.formValues.StackFile)
    );
  }
}

/**
 *
 * @param {'template'|'repository' | 'editor' | 'upload'} method
 * @param {import('@/react/portainer/templates/custom-templates/types').CustomTemplate | undefined} template
 * @returns 'repository' | 'editor' | 'upload'
 */
function getMethod(method, template) {
  if (method !== 'template') {
    return method;
  }

  if (template && template.GitConfig) {
    return 'repository';
  }
  return 'editor';
}
