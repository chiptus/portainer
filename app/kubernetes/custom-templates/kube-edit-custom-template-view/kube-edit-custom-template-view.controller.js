import { ResourceControlViewModel } from '@/react/portainer/access-control/models/ResourceControlViewModel';
import { AccessControlFormData } from '@/portainer/components/accessControlForm/porAccessControlFormModel';
import { isBE } from '@/react/portainer/feature-flags/feature-flags.service';
import { getFilePreview } from '@/react/portainer/gitops/gitops.service';
import { getTemplateVariables, intersectVariables } from '@/react/portainer/custom-templates/components/utils';
import { getDeploymentOptions } from '@/react/portainer/environments/environment.service';
import { confirmWebEditorDiscard } from '@@/modals/confirm';
import { KUBE_TEMPLATE_NAME_VALIDATION_REGEX } from '@/constants';

class KubeEditCustomTemplateViewController {
  /* @ngInject */
  constructor($async, $state, EndpointProvider, Authentication, CustomTemplateService, FormValidator, Notifications, ResourceControlService) {
    Object.assign(this, { $async, $state, EndpointProvider, Authentication, CustomTemplateService, FormValidator, Notifications, ResourceControlService });

    this.isTemplateVariablesEnabled = isBE;

    this.formValues = {
      Variables: [],
      TLSSkipVerify: false,
      Title: '',
      Description: '',
      Note: '',
      Logo: '',
    };
    this.state = {
      formValidationError: '',
      isEditorDirty: false,
      isTemplateValid: true,
      isLoading: true,
      isEditorReadOnly: false,
      templateLoadFailed: false,
      templatePreviewFailed: false,
      templatePreviewError: '',
    };
    this.templates = [];

    this.validationData = {
      title: {
        pattern: KUBE_TEMPLATE_NAME_VALIDATION_REGEX,
        error:
          "This field must consist of lower-case alphanumeric characters, '.', '_' or '-', must start and end with an alphanumeric character and must be 63 characters or less (e.g. 'my-name', or 'abc-123').",
      },
    };

    this.getTemplate = this.getTemplate.bind(this);
    this.submitAction = this.submitAction.bind(this);
    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.onBeforeUnload = this.onBeforeUnload.bind(this);
    this.handleChange = this.handleChange.bind(this);
    this.onVariablesChange = this.onVariablesChange.bind(this);
    this.onChangeGitCredential = this.onChangeGitCredential.bind(this);
    this.previewFileFromGitRepository = this.previewFileFromGitRepository.bind(this);
    this.onChangePlatform = this.onChangePlatform.bind(this);
    this.onChangeType = this.onChangeType.bind(this);
  }

  onChangePlatform(value) {
    this.handleChange({ Platform: value });
  }

  onChangeType(value) {
    this.handleChange({ Type: value });
  }

  getTemplate() {
    return this.$async(async () => {
      try {
        const { id } = this.$state.params;

        const template = await this.CustomTemplateService.customTemplate(id);

        if (template.GitConfig !== null) {
          this.state.isEditorReadOnly = true;
        }

        try {
          template.FileContent = await this.CustomTemplateService.customTemplateFile(id, template.GitConfig !== null);
        } catch (err) {
          this.state.templateLoadFailed = true;
          throw err;
        }

        template.Variables = template.Variables || [];

        this.formValues = { ...this.formValues, ...template };

        this.parseTemplate(template.FileContent);
        this.parseGitConfig(template.GitConfig);

        this.oldFileContent = this.formValues.FileContent;

        this.formValues.ResourceControl = new ResourceControlViewModel(template.ResourceControl);
        this.formValues.AccessControlData = new AccessControlFormData();
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve custom template data');
      }
    });
  }

  onVariablesChange(values) {
    this.handleChange({ Variables: values });
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
      TLSSkipVerify: this.formValues.TLSSkipVerify,
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

  validateForm() {
    this.state.formValidationError = '';

    if (!this.formValues.FileContent) {
      this.state.formValidationError = 'Template file content must not be empty';
      return false;
    }

    const title = this.formValues.Title;
    const id = this.$state.params.id;

    const isNotUnique = this.templates.some((template) => template.Title === title && template.Id != id);
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
    return this.$async(async () => {
      if (!this.validateForm()) {
        return;
      }

      this.actionInProgress = true;
      try {
        await this.CustomTemplateService.updateCustomTemplate(this.formValues.Id, this.formValues);

        const userDetails = this.Authentication.getUserDetails();
        const userId = userDetails.ID;

        await this.ResourceControlService.applyResourceControl(userId, this.formValues.AccessControlData, this.formValues.ResourceControl);

        this.Notifications.success('Success', 'Custom template successfully updated');
        this.state.isEditorDirty = false;
        this.$state.go('kubernetes.templates.custom');
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to update custom template');
      } finally {
        this.actionInProgress = false;
      }
    });
  }

  onChangeFileContent(value) {
    if (stripSpaces(this.formValues.FileContent) !== stripSpaces(value)) {
      this.formValues.FileContent = value;
      this.parseTemplate(value);
      this.state.isEditorDirty = true;
    }
  }

  async $onInit() {
    this.$async(async () => {
      await this.getTemplate();

      try {
        const endpoint = this.EndpointProvider.currentEndpoint();
        this.deploymentOptions = await getDeploymentOptions(endpoint.Id);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve deployment options');
      }

      try {
        this.templates = await this.CustomTemplateService.customTemplates();
      } catch (err) {
        this.Notifications.error('Failure loading', err, 'Failed loading custom templates');
      }
      if (this.formValues.GitConfig !== null) {
        this.formValues = {
          ...this.formValues,
          GitCredentials: [],
          SaveCredential: false,
          NewCredentialName: '',
          NewCredentialNameExist: false,
        };

        if (this.formValues.RepositoryGitCredentialID > 0) {
          this.formValues.SelectedGitCredential = this.formValues.GitCredentials.find((x) => x.id === this.formValues.RepositoryGitCredentialID);
          if (this.formValues.SelectedGitCredential) {
            this.formValues.RepositoryGitCredentialID = this.formValues.SelectedGitCredential.id;
            this.formValues.RepositoryUsername = this.formValues.SelectedGitCredential.username;
            this.formValues.RepositoryPassword = '';
          }
        }
      }

      this.state.isLoading = false;

      window.addEventListener('beforeunload', this.onBeforeUnload);
    });
  }

  isEditorDirty() {
    return this.formValues.FileContent !== this.oldFileContent && this.state.isEditorDirty;
  }

  uiCanExit() {
    if (this.isEditorDirty()) {
      return confirmWebEditorDiscard();
    }
  }

  onBeforeUnload(event) {
    if (this.formValues.FileContent !== this.oldFileContent && this.state.isEditorDirty) {
      event.preventDefault();
      event.returnValue = '';

      return '';
    }
  }

  $onDestroy() {
    window.removeEventListener('beforeunload', this.onBeforeUnload);
  }
}

export default KubeEditCustomTemplateViewController;

function stripSpaces(str = '') {
  return str.replace(/(\r\n|\n|\r)/gm, '');
}
