import { EditorType } from '@/react/edge/edge-stacks/types';
import { getValidEditorTypes } from '@/react/edge/edge-stacks/utils';
import { STACK_NAME_VALIDATION_REGEX } from '@/react/constants';
import { confirmWebEditorDiscard } from '@@/modals/confirm';
import { baseEdgeStackWebhookUrl, createWebhookId } from '@/portainer/helpers/webhookHelper';
import { parseAutoUpdateResponse, transformAutoUpdateViewModel } from '@/react/portainer/gitops/AutoUpdateFieldset/utils';
import { EnvironmentType } from '@/react/portainer/environments/types';

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
      Registries: [],
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
  }

  onEnvVarChange(envVars) {
    return this.$scope.$evalAsync(() => {
      this.formValues.envVars = envVars;
    });
  }

  buildAnalyticsProperties() {
    const format = 'compose';
    const metadata = { type: methodLabel(this.state.Method), format };

    if (metadata.type === 'template') {
      metadata.templateName = this.selectedTemplate.title;
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

  async $onInit() {
    // Initial Registry Options for selector
    this.formValues.RegistryOptions = await this.RegistryService.registries();

    this.registryID = '';
    this.errorMessage = '';
    this.dryrun = false;

    try {
      this.edgeGroups = await this.EdgeGroupService.groups();
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve Edge groups');
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
    this.formValues.Registries = [];
    this.registryID = '';
    this.dryrun = false;
  }

  selectedRegistry(e) {
    return this.$async(async () => {
      const selectedRegistry = e;
      this.registryID = selectedRegistry;
      this.formValues.Registries = [this.registryID];
    });
  }

  matchRegistry() {
    return this.$async(async () => {
      const name = this.formValues.Name;
      this.state.actionInProgress = true;
      this.errorMessage = '';
      this.dryrun = true;
      let response = '';
      let method = this.state.Method;

      if (method === 'template') {
        method = 'editor';
      }

      if (method === 'editor' || method === 'upload') {
        try {
          if (method === 'editor') {
            response = await this.createStackFromFileContent(name, this.dryrun);
          }
          if (method === 'upload') {
            const responseFromUpload = await this.createStackFromFileUpload(name, this.dryrun);
            response = responseFromUpload.data;
          }
          if (response.Registries.length !== 0) {
            const validRegistry = this.checkRegistries(response.Registries);
            if (validRegistry) {
              this.registryID = response.Registries[0];
              this.formValues.Registries = [this.registryID];
            } else {
              this.registryID = '';
              this.errorMessage = ' Images need to be from a single registry, please edit and reload';
            }
          } else {
            this.registryID = '';
          }
        } catch (err) {
          this.Notifications.error('Failure', err, 'Unable to retrieve registries');
        } finally {
          this.dryrun = false;
          this.state.actionInProgress = false;
        }
      } else {
        // Git Repository does not has dryrun
        this.dryrun = false;
        this.state.actionInProgress = false;
      }
    });
  }

  createStack() {
    return this.$async(async () => {
      const name = this.formValues.Name;
      let method = this.state.Method;

      if (method === 'template') {
        method = 'editor';
      }

      if (!this.validateForm(method)) {
        return;
      }

      this.state.actionInProgress = true;
      try {
        await this.createStackByMethod(name, method, this.dryrun);

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

  createStackByMethod(name, method, dryrun) {
    switch (method) {
      case 'editor':
        return this.createStackFromFileContent(name, dryrun);
      case 'upload':
        return this.createStackFromFileUpload(name, dryrun);
      case 'repository':
        return this.createStackFromGitRepository(name);
    }
  }

  createStackFromFileContent(name, dryrun) {
    const { StackFileContent, Groups, DeploymentType, Registries, UseManifestNamespaces, PrePullImage, RetryDeploy, webhookEnabled, envVars } = this.formValues;

    let webhookId = '';
    if (webhookEnabled) {
      webhookId = this.state.webhookId;
    }

    return this.EdgeStackService.createStackFromFileContent(
      {
        name,
        StackFileContent,
        EdgeGroups: Groups,
        DeploymentType,
        Registries: dryrun ? [] : Registries,
        UseManifestNamespaces,
        PrePullImage,
        RetryDeploy,
        webhook: webhookId,
        envVars,
      },
      dryrun
    );
  }

  createStackFromFileUpload(name, dryrun) {
    const { StackFile, Groups, DeploymentType, Registries, UseManifestNamespaces, PrePullImage, RetryDeploy, webhookEnabled, envVars } = this.formValues;
    let webhookId = '';
    if (webhookEnabled) {
      webhookId = this.state.webhookId;
    }

    return this.EdgeStackService.createStackFromFileUpload(
      {
        Name: name,
        EdgeGroups: Groups,
        DeploymentType,
        Registries: dryrun ? [] : Registries,
        UseManifestNamespaces,
        PrePullImage,
        RetryDeploy,
        webhook: webhookId,
        envVars,
      },
      StackFile,
      dryrun
    );
  }

  async createStackFromGitRepository(name) {
    const {
      Groups,
      DeploymentType,
      Registries,
      UseManifestNamespaces,
      PrePullImage,
      RetryDeploy,
      SupportRelativePath,
      FilesystemPath,
      SupportPerDeviceConfigs,
      PerDeviceConfigsMatchType,
      PerDeviceConfigsPath,
      envVars,
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
        Registries: Registries,
        UseManifestNamespaces,
        PrePullImage,
        RetryDeploy,
        AutoUpdate: autoUpdate,
        SupportRelativePath,
        FilesystemPath,
        SupportPerDeviceConfigs,
        PerDeviceConfigsMatchType,
        PerDeviceConfigsPath,
        envVars,
      },
      repositoryOptions
    );
  }

  onChangeDeploymentType(deploymentType) {
    return this.$scope.$evalAsync(() => {
      this.formValues.DeploymentType = deploymentType;
      this.state.Method = 'editor';
      this.formValues.StackFileContent = '';
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
