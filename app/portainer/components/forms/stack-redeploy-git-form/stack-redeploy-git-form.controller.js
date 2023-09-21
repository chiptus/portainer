import { RepositoryMechanismTypes } from 'Kubernetes/models/deploy';
import { confirmStackUpdate } from '@/react/common/stacks/common/confirm-stack-update';
import { parseAutoUpdateResponse } from '@/react/portainer/gitops/AutoUpdateFieldset/utils';
import { baseStackWebhookUrl, createWebhookId } from '@/portainer/helpers/webhookHelper';
import { confirmEnableTLSVerify } from '@/react/portainer/gitops/utils';

class StackRedeployGitFormController {
  /* @ngInject */
  constructor($async, $state, StackService, UserService, Authentication, Notifications, FormHelper) {
    this.$async = $async;
    this.$state = $state;
    this.StackService = StackService;
    this.UserService = UserService;
    this.Authentication = Authentication;
    this.Notifications = Notifications;
    this.FormHelper = FormHelper;

    this.state = {
      inProgress: false,
      redeployInProgress: false,
      showConfig: false,
      isEdit: false,

      // isAuthEdit is used to preserve the editing state of the AuthFieldset component.
      // Within the stack editing page, users have the option to turn the AuthFieldset on or off
      // and save the stack setting. If the user enables the AuthFieldset from off to on, it
      // implies that they must input new Git authentication, rather than edit existing authentication.
      // Thus, a dedicated state tracker is required to differentiate between the editing state of
      // AuthFieldset component and the parent Stack component
      // When isAuthEdit is true, PAT field needs to be validated.
      isAuthEdit: false,
      hasUnsavedChanges: false,
      baseWebhookUrl: baseStackWebhookUrl(),
      webhookId: createWebhookId(),
    };

    this.formValues = {
      RepositoryURL: '',
      RepositoryURLValid: true,
      RefName: '',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      RepositoryGitCredentialID: 0,
      SaveCredential: true,
      NewCredentialName: '',

      Env: [],
      Option: {
        Prune: false,
      },
      PullImage: false,
      // auto update
      AutoUpdate: parseAutoUpdateResponse(),
    };

    this.onChange = this.onChange.bind(this);
    this.onChangeRef = this.onChangeRef.bind(this);
    this.onChangeAutoUpdate = this.onChangeAutoUpdate.bind(this);
    this.onChangeEnvVar = this.onChangeEnvVar.bind(this);
    this.onChangeOption = this.onChangeOption.bind(this);
    this.onChangeGitAuth = this.onChangeGitAuth.bind(this);
    this.onChangeTLSSkipVerify = this.onChangeTLSSkipVerify.bind(this);
  }

  buildAnalyticsProperties() {
    const metadata = {};

    if (this.formValues.RepositoryAutomaticUpdates) {
      metadata.automaticUpdates = autoSyncLabel(this.formValues.RepositoryMechanism);
    }
    return { metadata };

    function autoSyncLabel(type) {
      switch (type) {
        case RepositoryMechanismTypes.INTERVAL:
          return 'polling';
        case RepositoryMechanismTypes.WEBHOOK:
          return 'webhook';
      }
      return 'off';
    }
  }

  onChange(values) {
    return this.$async(async () => {
      this.formValues = {
        ...this.formValues,
        ...values,
      };
      this.state.hasUnsavedChanges = angular.toJson(this.savedFormValues) !== angular.toJson(this.formValues);
    });
  }

  onChangeRef(value) {
    this.onChange({ RefName: value });
  }

  onChangeEnvVar(value) {
    this.onChange({ Env: value });
  }

  async onChangeTLSSkipVerify(value) {
    return this.$async(async () => {
      if (this.model.TLSSkipVerify && !value) {
        const confirmed = await confirmEnableTLSVerify();

        if (!confirmed) {
          return;
        }
      }
      this.onChange({ TLSSkipVerify: value });
    });
  }

  onChangeOption(values) {
    this.onChange({
      Option: {
        Prune: values,
      },
    });
  }

  async submit() {
    const isSwarmStack = this.stack.Type === 1;
    const that = this;
    confirmStackUpdate(
      'Any changes to this stack or application made locally in Portainer will be overridden, which may cause service interruption. Do you wish to continue?',
      isSwarmStack
    ).then(async function (result) {
      if (!result) {
        return;
      }
      try {
        that.state.redeployInProgress = true;

        // save git credential
        if (that.formValues.SaveCredential && that.formValues.NewCredentialName) {
          const userDetails = that.Authentication.getUserDetails();
          await that.UserService.saveGitCredential(userDetails.ID, that.formValues.NewCredentialName, that.formValues.RepositoryUsername, that.formValues.RepositoryPassword).then(
            function success(data) {
              that.formValues.RepositoryGitCredentialID = data.gitCredential.id;
            }
          );
        }

        await that.StackService.updateGit(
          that.stack.Id,
          that.stack.EndpointId,
          that.FormHelper.removeInvalidEnvVars(that.formValues.Env),
          that.formValues.Option.Prune,
          that.formValues,
          result.pullImage
        );
        that.Notifications.success('Success', 'Pulled and redeployed stack successfully');
        that.$state.reload();
      } catch (err) {
        that.Notifications.error('Failure', err, 'Failed redeploying stack');
      } finally {
        that.state.redeployInProgress = false;
      }
    });
  }

  async saveGitSettings() {
    const that = this;
    const userDetails = this.Authentication.getUserDetails();
    return this.$async(async () => {
      try {
        this.state.inProgress = true;

        // save git credential
        if (this.formValues.SaveCredential && this.formValues.NewCredentialName) {
          await this.UserService.saveGitCredential(userDetails.ID, this.formValues.NewCredentialName, this.formValues.RepositoryUsername, this.formValues.RepositoryPassword).then(
            function success(data) {
              that.formValues = {
                ...that.formValues,
                RepositoryGitCredentialID: data.gitCredential.id,
                SaveCredential: false,
                RepositoryPassword: '',
              };
            }
          );
        }

        const stack = await this.StackService.updateGitStackSettings(
          this.stack.Id,
          this.stack.EndpointId,
          this.FormHelper.removeInvalidEnvVars(this.formValues.Env),
          this.formValues,
          this.state.webhookId
        );

        this.savedFormValues = angular.copy(this.formValues);
        this.state.hasUnsavedChanges = false;
        this.Notifications.success('Success', 'Save stack settings successfully');

        if (!(this.stack.GitConfig && this.stack.GitConfig.Authentication)) {
          // update the AuthFieldset setting
          this.state.isAuthEdit = false;
          this.formValues.RepositoryUsername = '';
          this.formValues.RepositoryPassword = '';
        }
        this.stack = stack;
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to save stack settings');
        if (that.formValues.SaveCredential && that.formValues.NewCredentialName && that.formValues.RepositoryGitCredentialID) {
          that.UserService.deleteGitCredential(userDetails.ID, that.formValues.RepositoryGitCredentialID);
        }
      } finally {
        this.state.inProgress = false;
      }
    });
  }

  disablePullAndRedeployButton() {
    return this.isSubmitButtonDisabled() || this.state.hasUnsavedChanges || !this.redeployGitForm.$valid;
  }

  disableSaveSettingsButton() {
    return (
      this.isSubmitButtonDisabled() ||
      !this.state.hasUnsavedChanges ||
      !this.redeployGitForm.$valid ||
      (this.formValues.RepositoryAuthentication && !this.formValues.RepositoryPassword && this.formValues.RepositoryGitCredentialID === 0 && !this.formValues.StackID === 0) ||
      (this.formValues.RepositoryAuthentication &&
        this.formValues.RepositoryPassword &&
        this.formValues.SaveCredential &&
        (!this.formValues.NewCredentialName || this.formValues.NewCredentialNameExist || this.formValues.NewCredentialNameInvalid))
    );
  }

  isSubmitButtonDisabled() {
    return this.state.inProgress || this.state.redeployInProgress;
  }

  isAutoUpdateChanged() {
    const wasEnabled = !!(this.stack.AutoUpdate && (this.stack.AutoUpdate.Interval || this.stack.AutoUpdate.Webhook));
    const isEnabled = this.formValues.AutoUpdate.RepositoryAutomaticUpdates;
    return isEnabled !== wasEnabled;
  }

  onChangeGitAuth(values) {
    this.onChange(values);
  }

  onChangeAutoUpdate(values) {
    this.onChange({
      AutoUpdate: {
        ...this.formValues.AutoUpdate,
        ...values,
      },
    });
  }

  async $onInit() {
    this.formValues.RepositoryURL = this.model.URL;
    this.formValues.RefName = this.model.ReferenceName;
    this.formValues.TLSSkipVerify = this.model.TLSSkipVerify;
    this.formValues.Env = this.stack.Env;

    if (this.stack.Option) {
      this.formValues.Option = this.stack.Option;
    }

    this.formValues.AutoUpdate = parseAutoUpdateResponse(this.stack.AutoUpdate);

    if (this.stack.AutoUpdate && this.stack.AutoUpdate.Webhook) {
      this.state.webhookId = this.stack.AutoUpdate.Webhook;
    }

    if (this.stack.GitConfig && this.stack.GitConfig.Authentication) {
      this.formValues.RepositoryUsername = this.stack.GitConfig.Authentication.Username;
      this.formValues.RepositoryPassword = this.stack.GitConfig.Authentication.Password;
      this.formValues.RepositoryAuthentication = true;
      this.state.isEdit = true;
      this.state.isAuthEdit = true;

      if (this.stack.GitConfig.Authentication.GitCredentialID > 0) {
        this.formValues.SaveCredential = false;
        this.formValues.RepositoryGitCredentialID = this.stack.GitConfig.Authentication.GitCredentialID;
      }
      this.gitStackId = this.stack.Id;
    }
    this.savedFormValues = angular.copy(this.formValues);
  }
}

export default StackRedeployGitFormController;
