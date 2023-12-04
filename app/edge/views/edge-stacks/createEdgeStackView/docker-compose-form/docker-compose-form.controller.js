import { getInitialTemplateValues } from '@/react/edge/edge-stacks/CreateView/TemplateFieldset';
import { editor, git, edgeStackTemplate, upload } from '@@/BoxSelector/common-options/build-methods';

class DockerComposeFormController {
  /* @ngInject */
  constructor($async, Notifications, UserService, Authentication) {
    Object.assign(this, { $async, Notifications, UserService, Authentication });

    this.methodOptions = [editor, upload, git, edgeStackTemplate];

    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.onChangeFile = this.onChangeFile.bind(this);
    this.onChangeMethod = this.onChangeMethod.bind(this);
    this.onChangeFormValues = this.onChangeFormValues.bind(this);
    this.onEnableRelativePathsChange = this.onEnableRelativePathsChange.bind(this);
    this.onEnablePerDeviceConfigsChange = this.onEnablePerDeviceConfigsChange.bind(this);
    this.onPerDeviceConfigsPathChange = this.onPerDeviceConfigsPathChange.bind(this);
    this.isGitTemplate = this.isGitTemplate.bind(this);
  }

  isGitTemplate() {
    return this.state.Method === 'template' && !!this.templateValues.template && !!this.templateValues.template.GitConfig;
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
    this.setTemplateValues(getInitialTemplateValues());
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
      this.formValues.PerDeviceConfigsPath = '';
      this.formValues.SupportPerDeviceConfigs = false;
      this.formValues.PerDeviceConfigsMatchType = '';
      this.formValues.PerDeviceConfigsGroupMatchType = '';
    });
  }
}

export default DockerComposeFormController;
