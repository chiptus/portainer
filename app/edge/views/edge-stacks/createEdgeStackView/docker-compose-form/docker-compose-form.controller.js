import { editor, git, edgeStackTemplate, upload } from '@@/BoxSelector/common-options/build-methods';

import { EdgeIdInsightsBoxContent } from './EdgeIdInsightsBoxContent';

class DockerComposeFormController {
  /* @ngInject */
  constructor($async, EdgeTemplateService, Notifications, UserService, Authentication) {
    Object.assign(this, { $async, EdgeTemplateService, Notifications, UserService, Authentication });

    this.methodOptions = [editor, upload, git, edgeStackTemplate];

    this.selectedTemplate = null;

    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.onChangeFile = this.onChangeFile.bind(this);
    this.onChangeTemplate = this.onChangeTemplate.bind(this);
    this.onChangeMethod = this.onChangeMethod.bind(this);
    this.onChangeFormValues = this.onChangeFormValues.bind(this);
    this.onEnableRelativePathsChange = this.onEnableRelativePathsChange.bind(this);
    this.onEnablePerDeviceConfigsChange = this.onEnablePerDeviceConfigsChange.bind(this);
    this.onPerDeviceConfigsPathChange = this.onPerDeviceConfigsPathChange.bind(this);
  }

  onChangeFormValues(newValues) {
    return this.$async(async () => {
      this.formValues = {
        ...this.formValues,
        ...newValues,
      };
    });
  }

  onChangeMethod(method) {
    this.state.Method = method;
    this.formValues.StackFileContent = '';
    this.selectedTemplate = null;
  }

  onChangeTemplate(template) {
    return this.$async(async () => {
      this.formValues.StackFileContent = '';
      try {
        const fileContent = await this.EdgeTemplateService.edgeTemplate(template);
        this.formValues.StackFileContent = fileContent;
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve Template');
      }
    });
  }

  onChangeFileContent(value) {
    return this.$async(async () => {
      this.formValues.StackFileContent = value;
    });
  }

  onChangeFile(value) {
    return this.$async(async () => {
      this.formValues.StackFile = value;
    });
  }

  onEnableRelativePathsChange(value) {
    return this.$async(async () => {
      this.formValues.SupportRelativePath = value;
    });
  }

  onEnablePerDeviceConfigsChange(value) {
    return this.$async(async () => {
      this.formValues.SupportPerDeviceConfigs = value;
    });
  }

  onPerDeviceConfigsPathChange(value) {
    return this.$async(async () => {
      this.formValues.PerDeviceConfigsPath = value;
    });
  }

  async $onInit() {
    return this.$async(async () => {
      this.formValues.SupportPerDeviceConfigs = false;
      this.formValues.PerDeviceConfigsMatchType = 'file';
      this.edgeIdInsightsBoxContent = EdgeIdInsightsBoxContent();

      try {
        const templates = await this.EdgeTemplateService.edgeTemplates();
        this.templates = templates.map((template) => ({ ...template, label: `${template.title} - ${template.description}` }));
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve Templates');
      }
    });
  }
}

export default DockerComposeFormController;
