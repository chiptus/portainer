import { EditorType } from '@/edge/types';
import { PortainerEndpointTypes } from 'Portainer/models/endpoint/models';
import { getValidEditorTypes } from '@/edge/utils';

export default class CreateEdgeStackViewController {
  /* @ngInject */
  constructor($state, $window, ModalService, EdgeStackService, EdgeGroupService, Notifications, $async, $scope, RegistryService) {
    Object.assign(this, { $state, $window, ModalService, EdgeStackService, EdgeGroupService, Notifications, $async, RegistryService, $scope });

    this.formValues = {
      Name: '',
      StackFileContent: '',
      StackFile: null,
      RepositoryURL: '',
      RepositoryReferenceName: '',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      Env: [],
      ComposeFilePathInRepository: '',
      Groups: [],
      DeploymentType: EditorType.Compose,
      RegistryOptions: [],
      Registries: [],
    };

    this.state = {
      Method: 'editor',
      formValidationError: '',
      actionInProgress: false,
      StackType: null,
      isEditorDirty: false,
      endpointTypes: [],
    };

    this.edgeGroups = null;

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
    this.hasDockerEndpoint = this.hasDockerEndpoint.bind(this);
    this.hasKubeEndpoint = this.hasKubeEndpoint.bind(this);
    this.hasNomadEndpoint = this.hasNomadEndpoint.bind(this);
    this.onChangeDeploymentType = this.onChangeDeploymentType.bind(this);
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
      return this.ModalService.confirmWebEditorDiscard();
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
      this.noGroups = this.edgeGroups.length === 0;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve Edge groups');
    }

    this.$window.onbeforeunload = () => {
      if (this.state.Method === 'editor' && this.formValues.StackFileContent && this.state.isEditorDirty) {
        return '';
      }
    };
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
      this.registryID = selectedRegistry.Id;
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
      } finally {
        this.state.actionInProgress = false;
      }
    });
  }

  onChangeGroups(groups) {
    this.formValues.Groups = groups;

    this.checkIfEndpointTypes(groups);
  }

  checkIfEndpointTypes(groups) {
    const edgeGroups = groups.map((id) => this.edgeGroups.find((e) => e.Id === id));
    this.state.endpointTypes = edgeGroups.flatMap((group) => group.EndpointTypes);
    this.selectValidDeploymentType();
  }

  selectValidDeploymentType() {
    const validTypes = getValidEditorTypes(this.state.endpointTypes);

    if (!validTypes.includes(this.formValues.DeploymentType)) {
      this.onChangeDeploymentType(validTypes[0]);
    }
  }

  hasKubeEndpoint() {
    return this.state.endpointTypes.includes(PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment);
  }

  hasDockerEndpoint() {
    return this.state.endpointTypes.includes(PortainerEndpointTypes.EdgeAgentOnDockerEnvironment);
  }

  hasNomadEndpoint() {
    return this.state.endpointTypes.includes(PortainerEndpointTypes.EdgeAgentOnNomadEnvironment);
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
    const { StackFileContent, Groups, DeploymentType, Registries } = this.formValues;
    return this.EdgeStackService.createStackFromFileContent(
      {
        name,
        StackFileContent,
        EdgeGroups: Groups,
        DeploymentType,
        Registries: dryrun ? [] : Registries,
      },
      dryrun
    );
  }

  createStackFromFileUpload(name, dryrun) {
    const { StackFile, Groups, DeploymentType, Registries } = this.formValues;
    return this.EdgeStackService.createStackFromFileUpload(
      {
        Name: name,
        EdgeGroups: Groups,
        DeploymentType,
        Registries: dryrun ? [] : Registries,
      },
      StackFile,
      dryrun
    );
  }

  createStackFromGitRepository(name) {
    const { Groups, DeploymentType, Registries } = this.formValues;
    const repositoryOptions = {
      RepositoryURL: this.formValues.RepositoryURL,
      RepositoryReferenceName: this.formValues.RepositoryReferenceName,
      FilePathInRepository: this.formValues.ComposeFilePathInRepository,
      RepositoryAuthentication: this.formValues.RepositoryAuthentication,
      RepositoryUsername: this.formValues.RepositoryUsername,
      RepositoryPassword: this.formValues.RepositoryPassword,
    };

    return this.EdgeStackService.createStackFromGitRepository(
      {
        name,
        EdgeGroups: Groups,
        DeploymentType,
        Registries: Registries,
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

  formIsInvalid() {
    return (
      this.form.$invalid ||
      !this.formValues.Groups.length ||
      (['template', 'editor'].includes(this.state.Method) && !this.formValues.StackFileContent) ||
      ('upload' === this.state.Method && !this.formValues.StackFile)
    );
  }
}
