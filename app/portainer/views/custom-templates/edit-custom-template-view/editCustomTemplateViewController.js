import _ from 'lodash';
import { getFilePreview } from '@/react/portainer/gitops/gitops.service';
import { ResourceControlViewModel } from '@/react/portainer/access-control/models/ResourceControlViewModel';
import { TEMPLATE_NAME_VALIDATION_REGEX } from '@/constants';

import { AccessControlFormData } from 'Portainer/components/accessControlForm/porAccessControlFormModel';
import { getTemplateVariables, intersectVariables } from '@/react/portainer/custom-templates/components/utils';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';
import { confirmWebEditorDiscard } from '@@/modals/confirm';

class EditCustomTemplateViewController {
  /* @ngInject */
  constructor($async, $state, $window, Authentication, CustomTemplateService, FormValidator, Notifications, ResourceControlService, EndpointProvider) {
    Object.assign(this, { $async, $state, $window, Authentication, CustomTemplateService, FormValidator, Notifications, ResourceControlService, EndpointProvider });

    this.isTemplateVariablesEnabled = isBE;

    this.formValues = {
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
      Variables: [],
      TLSSkipVerify: false,
    };
    this.state = {
      formValidationError: '',
      isEditorDirty: false,
      isTemplateValid: true,
      isLoading: true,
      isEditorVisible: false,
      isEditorReadOnly: false,
      templateLoadFailed: false,
      templatePreviewFailed: false,
      templatePreviewError: '',
      templateNameRegex: TEMPLATE_NAME_VALIDATION_REGEX,
    };
    this.templates = [];

    this.getTemplate = this.getTemplate.bind(this);
    this.getTemplateAsync = this.getTemplateAsync.bind(this);
    this.submitAction = this.submitAction.bind(this);
    this.submitActionAsync = this.submitActionAsync.bind(this);
    this.editorUpdate = this.editorUpdate.bind(this);
    this.onVariablesChange = this.onVariablesChange.bind(this);
    this.handleChange = this.handleChange.bind(this);
    this.previewFileFromGitRepository = this.previewFileFromGitRepository.bind(this);
  }

  getTemplate() {
    return this.$async(this.getTemplateAsync);
  }
  async getTemplateAsync() {
    try {
      const template = await this.CustomTemplateService.customTemplate(this.$state.params.id);

      if (template.GitConfig !== null) {
        this.state.isEditorReadOnly = true;
      }

      try {
        template.FileContent = await this.CustomTemplateService.customTemplateFile(this.$state.params.id, template.GitConfig !== null);
      } catch (err) {
        this.state.templateLoadFailed = true;
        throw err;
      }

      template.Variables = template.Variables || [];

      this.formValues = { ...this.formValues, ...template };

      this.parseTemplate(template.FileContent);
      this.parseGitConfig(template.GitConfig);

      this.oldFileContent = this.formValues.FileContent;
      if (template.ResourceControl) {
        this.formValues.ResourceControl = new ResourceControlViewModel(template.ResourceControl);
      }
      this.formValues.AccessControlData = new AccessControlFormData();
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve custom template data');
    }
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

      if (this.formValues.GitCredentials) {
        const existGitCredential = this.formValues.GitCredentials.find((x) => x.name === this.formValues.NewCredentialName);
        this.formValues.NewCredentialNameExist = existGitCredential ? true : false;
        this.formValues.NewCredentialNameInvalid = !!(this.formValues.NewCredentialName && !this.formValues.NewCredentialName.match(/^[-_a-z0-9]+$/));
      }
    });
  }

  onChangeGitCredential(selectedGitCredential) {
    return this.$async(async () => {
      if (selectedGitCredential) {
        this.formValues.SelectedGitCredential = selectedGitCredential;
        this.formValues.RepositoryGitCredentialID = Number(selectedGitCredential.id);
        this.formValues.RepositoryUsername = selectedGitCredential.username;
        this.formValues.RepositoryPassword = '';
        this.formValues.SaveCredential = false;
        this.formValues.NewCredentialName = '';
      } else {
        this.formValues.SelectedGitCredential = null;
        this.formValues.RepositoryUsername = '';
        this.formValues.RepositoryPassword = '';
        this.formValues.RepositoryGitCredentialID = 0;
      }
    });
  }

  validateForm() {
    this.state.formValidationError = '';

    if (!this.formValues.FileContent) {
      this.state.formValidationError = 'Template file content must not be empty';
      return false;
    }

    const title = this.formValues.Title;
    const id = this.$state.params.id;
    const isNotUnique = _.some(this.templates, (template) => template.Title === title && template.Id != id);
    if (isNotUnique) {
      this.state.formValidationError = `A template with the name ${title} already exists`;
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

  submitAction() {
    return this.$async(this.submitActionAsync);
  }
  async submitActionAsync() {
    if (!this.validateForm()) {
      return;
    }

    this.actionInProgress = true;
    try {
      await this.CustomTemplateService.updateCustomTemplate(this.formValues.Id, this.formValues);

      const userDetails = this.Authentication.getUserDetails();
      // save git credential
      if (this.formValues.SaveCredential && this.formValues.NewCredentialName) {
        const data = await this.UserService.saveGitCredential(
          userDetails.ID,
          this.formValues.NewCredentialName,
          this.formValues.RepositoryUsername,
          this.formValues.RepositoryPassword
        );

        this.formValues.RepositoryGitCredentialID = data.gitCredential.id;
      }
      const userId = userDetails.ID;
      await this.ResourceControlService.applyResourceControl(userId, this.formValues.AccessControlData, this.formValues.ResourceControl);

      this.Notifications.success('Success', 'Custom template successfully updated');
      this.state.isEditorDirty = false;
      this.$state.go('docker.templates.custom');
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to update custom template');
    } finally {
      this.actionInProgress = false;
    }
  }

  editorUpdate(value) {
    if (this.formValues.FileContent.replace(/(\r\n|\n|\r)/gm, '') !== value.replace(/(\r\n|\n|\r)/gm, '')) {
      this.formValues.FileContent = value;
      this.parseTemplate(value);
      this.state.isEditorDirty = true;
    }
  }

  parseTemplate(templateStr) {
    if (!this.isTemplateVariablesEnabled) {
      return;
    }

    const variables = getTemplateVariables(templateStr);

    const isValid = !!variables;

    this.state.isTemplateValid = isValid;

    if (isValid) {
      this.onVariablesChange(variables.length > 0 ? intersectVariables(this.formValues.Variables, variables) : variables);
    }
  }

  parseGitConfig(config) {
    if (config === null) {
      return;
    }

    let flatConfig = {
      RepositoryURL: config.URL,
      RepositoryReferenceName: config.ReferenceName,
      ComposeFilePathInRepository: config.ConfigFilePath,
      RepositoryAuthentication: config.Authentication !== null,
      TLSSkipVerify: config.TLSSkipVerify,
    };

    if (config.Authentication) {
      flatConfig = {
        ...flatConfig,
        RepositoryUsername: config.Authentication.Username,
        RepositoryPassword: config.Authentication.Password,
        RepositoryGitCredentialID: config.Authentication.GitCredentialID,
        SaveCredential: false,
      };
    }

    this.formValues = { ...this.formValues, ...flatConfig };
  }

  previewFileFromGitRepository() {
    this.state.templatePreviewFailed = false;
    this.state.templatePreviewError = '';

    let creds = {};
    if (this.formValues.RepositoryAuthentication) {
      if (this.formValues.RepositoryPassword) {
        creds = {
          username: this.formValues.RepositoryUsername,
          password: this.formValues.RepositoryPassword,
        };
      } else if (this.formValues.SelectedGitCredential) {
        creds = { gitCredentialId: this.formValues.SelectedGitCredential.id };
      }
    }
    const payload = {
      repository: this.formValues.RepositoryURL,
      targetFile: this.formValues.ComposeFilePathInRepository,
      tlsSkipVerify: this.formValues.TLSSkipVerify,
      ...creds,
    };

    this.$async(async () => {
      try {
        this.formValues.FileContent = await getFilePreview(payload);
        this.state.isEditorDirty = true;

        // check if the template contains mustache template symbol
        this.parseTemplate(this.formValues.FileContent);
      } catch (err) {
        this.state.templatePreviewError = err.message;
        this.state.templatePreviewFailed = true;
      }
    });
  }

  async uiCanExit() {
    if (this.formValues.FileContent !== this.oldFileContent && this.state.isEditorDirty) {
      return confirmWebEditorDiscard();
    }
  }

  async $onInit() {
    return this.$async(async () => {
      await this.getTemplate();

      try {
        this.templates = await this.CustomTemplateService.customTemplates([1, 2]);
      } catch (err) {
        this.Notifications.error('Failure loading', err, 'Failed loading custom templates');
      }

      try {
        const endpoint = await this.EndpointProvider.currentEndpoint();
        this.deploymentOptions = await getDeploymentOptions(endpoint.Id);
      } catch (err) {
        this.Notifications.error('Failure loading', err, 'Failed loading deployment options');
      }

      this.state.isLoading = false;

      this.$window.onbeforeunload = () => {
        if (this.formValues.FileContent !== this.oldFileContent && this.state.isEditorDirty) {
          return '';
        }
      };
    });
  }

  $onDestroy() {
    this.state.isEditorDirty = false;
  }
}

export default EditCustomTemplateViewController;
