import _ from 'lodash';
import { AccessControlFormData } from 'Portainer/components/accessControlForm/porAccessControlFormModel';
import { TEMPLATE_NAME_VALIDATION_REGEX } from '@/constants';
import { getTemplateVariables, intersectVariables } from '@/react/portainer/custom-templates/components/utils';
import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { editor, upload, git } from '@@/BoxSelector/common-options/build-methods';
import { EnvironmentType } from '@/react/portainer/environments/types';
import { confirmWebEditorDiscard } from '@@/modals/confirm';

class CreateCustomTemplateViewController {
  /* @ngInject */
  constructor(
    $async,
    $state,
    $scope,
    $window,
    EndpointProvider,
    Authentication,
    CustomTemplateService,
    FormValidator,
    Notifications,
    ResourceControlService,
    StackService,
    StateManager,
    UserService
  ) {
    Object.assign(this, {
      $async,
      $state,
      $window,
      $scope,
      EndpointProvider,
      Authentication,
      CustomTemplateService,
      FormValidator,
      Notifications,
      ResourceControlService,
      StackService,
      StateManager,
      UserService,
    });

    this.buildMethods = [editor, upload, git];

    this.isTemplateVariablesEnabled = isBE;

    this.formValues = {
      Title: '',
      FileContent: '',
      File: null,
      RepositoryURL: '',
      RepositoryURLValid: false,
      RepositoryReferenceName: 'refs/heads/main',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      SaveCredential: true,
      RepositoryGitCredentialID: 0,
      NewCredentialName: '',
      NewCredentialNameExist: false,
      ComposeFilePathInRepository: 'docker-compose.yml',
      Description: '',
      Note: '',
      Logo: '',
      Platform: 1,
      Type: 1,
      AccessControlData: new AccessControlFormData(),
      Variables: [],
    };

    this.state = {
      Method: 'editor',
      formValidationError: '',
      actionInProgress: false,
      fromStack: false,
      loading: true,
      isEditorDirty: false,
      templateNameRegex: TEMPLATE_NAME_VALIDATION_REGEX,
      isTemplateValid: true,
    };

    this.templates = [];

    this.createCustomTemplate = this.createCustomTemplate.bind(this);
    this.createCustomTemplateAsync = this.createCustomTemplateAsync.bind(this);
    this.validateForm = this.validateForm.bind(this);
    this.createCustomTemplateByMethod = this.createCustomTemplateByMethod.bind(this);
    this.createCustomTemplateFromFileContent = this.createCustomTemplateFromFileContent.bind(this);
    this.createCustomTemplateFromFileUpload = this.createCustomTemplateFromFileUpload.bind(this);
    this.createCustomTemplateFromGitRepository = this.createCustomTemplateFromGitRepository.bind(this);
    this.editorUpdate = this.editorUpdate.bind(this);
    this.onChangeMethod = this.onChangeMethod.bind(this);
    this.onVariablesChange = this.onVariablesChange.bind(this);
    this.handleChange = this.handleChange.bind(this);
  }

  onVariablesChange(value) {
    this.handleChange({ Variables: value });
  }

  handleChange(values) {
    return this.$async(async () => {
      this.formValues = {
        ...this.formValues,
        ...values,
      };
    });
  }

  createCustomTemplate() {
    return this.$async(this.createCustomTemplateAsync);
  }

  onChangeMethod(method) {
    return this.$scope.$evalAsync(() => {
      this.formValues.FileContent = '';
      this.formValues.Variables = [];
      this.selectedTemplate = null;
      this.state.Method = method;
    });
  }

  async createCustomTemplateAsync() {
    let method = this.state.Method;

    if (method === 'template') {
      method = 'editor';
    }

    if (!this.validateForm(method)) {
      return;
    }

    this.state.actionInProgress = true;
    const userDetails = this.Authentication.getUserDetails();
    const that = this;
    try {
      if (method === 'repository') {
        // save git credential
        if (this.formValues.SaveCredential && this.formValues.NewCredentialName) {
          await this.UserService.saveGitCredential(userDetails.ID, this.formValues.NewCredentialName, this.formValues.RepositoryUsername, this.formValues.RepositoryPassword).then(
            function success(data) {
              that.formValues.RepositoryGitCredentialID = data.gitCredential.id;
            }
          );
        }
      }
      const customTemplate = await this.createCustomTemplateByMethod(method);

      const accessControlData = this.formValues.AccessControlData;

      const userId = userDetails.ID;
      await this.ResourceControlService.applyResourceControl(userId, accessControlData, customTemplate.ResourceControl);

      this.Notifications.success('Success', 'Custom template successfully created');
      this.state.isEditorDirty = false;
      this.$state.go('docker.templates.custom');
    } catch (err) {
      this.Notifications.error('Failure', err, 'A template with the same name already exists');
      const userDetails = this.Authentication.getUserDetails();
      if (this.formValues.SaveCredential && this.formValues.NewCredentialName && this.formValues.RepositoryGitCredentialID) {
        this.UserService.deleteGitCredential(userDetails.ID, this.formValues.RepositoryGitCredentialID);
      }
    } finally {
      this.state.actionInProgress = false;
    }
  }

  validateForm(method) {
    this.state.formValidationError = '';

    if (method === 'editor' && this.formValues.FileContent === '') {
      this.state.formValidationError = 'Template file content must not be empty';
      return false;
    }

    const title = this.formValues.Title;
    const isNotUnique = _.some(this.templates, (template) => template.Title === title);
    if (isNotUnique) {
      this.state.formValidationError = 'A template with the same name already exists';
      return false;
    }

    const isAdmin = this.Authentication.isAdmin();
    const accessControlData = this.formValues.AccessControlData;
    const error = this.FormValidator.validateAccessControl(accessControlData, isAdmin);

    if (error) {
      this.state.formValidationError = error;
      return false;
    }

    return true;
  }

  createCustomTemplateByMethod(method) {
    switch (method) {
      case 'editor':
        return this.createCustomTemplateFromFileContent();
      case 'upload':
        return this.createCustomTemplateFromFileUpload();
      case 'repository':
        return this.createCustomTemplateFromGitRepository();
    }
  }

  createCustomTemplateFromFileContent() {
    return this.CustomTemplateService.createCustomTemplateFromFileContent(this.formValues);
  }

  createCustomTemplateFromFileUpload() {
    return this.CustomTemplateService.createCustomTemplateFromFileUpload(this.formValues);
  }

  createCustomTemplateFromGitRepository() {
    return this.CustomTemplateService.createCustomTemplateFromGitRepository(this.formValues);
  }

  editorUpdate(value) {
    this.formValues.FileContent = value;
    this.state.isEditorDirty = true;
    this.parseTemplate(value);
  }

  parseTemplate(templateStr) {
    if (!this.isTemplateVariablesEnabled) {
      return;
    }

    const variables = getTemplateVariables(templateStr);

    const isValid = !!variables;

    this.state.isTemplateValid = isValid;

    if (isValid) {
      this.onVariablesChange(intersectVariables(this.formValues.Variables, variables));
    }
  }

  async $onInit() {
    const applicationState = this.StateManager.getState();

    this.state.endpointMode = applicationState.endpoint.mode;
    let stackType = 0;
    if (this.state.endpointMode.provider === 'DOCKER_STANDALONE') {
      this.isDockerStandalone = true;
      stackType = 2;
    } else if (this.state.endpointMode.provider === 'DOCKER_SWARM_MODE') {
      stackType = 1;
    }
    this.formValues.Type = stackType;

    const { fileContent, type } = this.$state.params;

    this.formValues.FileContent = fileContent;
    if (type) {
      this.formValues.Type = +type;
    }

    try {
      this.templates = await this.CustomTemplateService.customTemplates([1, 2]);
    } catch (err) {
      this.Notifications.error('Failure loading', err, 'Failed loading custom templates');
    }

    try {
      const endpoint = this.EndpointProvider.currentEndpoint();
      if (endpoint.Type === EnvironmentType.AgentOnKubernetes || endpoint.Type === EnvironmentType.EdgeAgentOnKubernetes || endpoint.Type === EnvironmentType.KubernetesLocal) {
        this.deploymentOptions = await getDeploymentOptions(endpoint.Id);
        this.methodOptions = [git];
        if (!this.deploymentOptions.hideWebEditor) {
          this.methodOptions.push(editor);
        }
        if (!this.deploymentOptions.hideFileUpload) {
          this.methodOptions.push(upload);
        }
        // the selected method must be available
        if (!this.methodOptions.map((option) => option.value).includes(this.state.Method)) {
          this.state.Method = this.methodOptions[0].value;
        }
      } else {
        this.methodOptions = [git, editor, upload];
      }
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to get deployment options');
    }

    this.state.loading = false;

    this.$window.onbeforeunload = () => {
      if (this.state.Method === 'editor' && this.formValues.FileContent && this.state.isEditorDirty) {
        return '';
      }
    };
  }

  $onDestroy() {
    this.state.isEditorDirty = false;
  }

  async uiCanExit() {
    if (this.state.Method === 'editor' && this.formValues.FileContent && this.state.isEditorDirty) {
      return confirmWebEditorDiscard();
    }
  }
}

export default CreateCustomTemplateViewController;
