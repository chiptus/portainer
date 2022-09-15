import uuidv4 from 'uuid/v4';
import { RepositoryMechanismTypes } from 'Kubernetes/models/deploy';
class StackRedeployGitFormController {
  /* @ngInject */
  constructor($async, $state, StackService, ModalService, UserService, Authentication, Notifications, WebhookHelper, FormHelper) {
    this.$async = $async;
    this.$state = $state;
    this.StackService = StackService;
    this.ModalService = ModalService;
    this.UserService = UserService;
    this.Authentication = Authentication;
    this.Notifications = Notifications;
    this.WebhookHelper = WebhookHelper;
    this.FormHelper = FormHelper;

    this.state = {
      inProgress: false,
      redeployInProgress: false,
      showConfig: false,
      isEdit: false,
      hasUnsavedChanges: false,
    };

    this.formValues = {
      RefName: '',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      RepositoryGitCredentialID: 0,
      SelectedGitCredential: null,
      GitCredentials: [],
      SaveCredential: true,
      NewCredentialName: '',
      NewCredentialNameExist: false,
      NewCredentialNameInvalid: false,
      Env: [],
      Option: {
        Prune: false,
      },
      PullImage: false,
      // auto update
      AutoUpdate: {
        RepositoryAutomaticUpdates: false,
        RepositoryAutomaticUpdatesForce: false,
        RepositoryMechanism: RepositoryMechanismTypes.INTERVAL,
        RepositoryFetchInterval: '5m',
        RepositoryWebhookURL: '',
        ForcePullImage: false,
      },
    };

    this.onChange = this.onChange.bind(this);
    this.onChangeRepositoryAuthentication = this.onChangeRepositoryAuthentication.bind(this);
    this.onChangeRepositoryUsername = this.onChangeRepositoryUsername.bind(this);
    this.onChangeRepositoryPassword = this.onChangeRepositoryPassword.bind(this);
    this.onSelectGitCredential = this.onSelectGitCredential.bind(this);
    this.onChangeSaveCredential = this.onChangeSaveCredential.bind(this);
    this.onChangeNewCredentialName = this.onChangeNewCredentialName.bind(this);
    this.onChangeRef = this.onChangeRef.bind(this);
    this.onChangeAutoUpdate = this.onChangeAutoUpdate.bind(this);
    this.onChangeEnvVar = this.onChangeEnvVar.bind(this);
    this.onChangeOption = this.onChangeOption.bind(this);
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
      const existGitCredential = this.formValues.GitCredentials.find((x) => x.name === this.formValues.NewCredentialName);
      this.formValues.NewCredentialNameExist = existGitCredential ? true : false;
      this.formValues.NewCredentialNameInvalid = this.formValues.NewCredentialName && !this.formValues.NewCredentialName.match(/^[-_a-z0-9]+$/) ? true : false;
      this.state.hasUnsavedChanges = angular.toJson(this.savedFormValues) !== angular.toJson(this.formValues);
    });
  }

  onChangeRef(value) {
    this.onChange({ RefName: value });
  }

  onChangeRepositoryAuthentication(value) {
    this.onChange({ RepositoryAuthentication: value });
  }

  onChangeRepositoryUsername(value) {
    this.onChange({ RepositoryUsername: value });
  }

  onChangeRepositoryPassword(value) {
    this.onChange({ RepositoryPassword: value });
  }

  onChangeSaveCredential(value) {
    this.onChange({ SaveCredential: value });
  }

  onChangeNewCredentialName(value) {
    this.onChange({ NewCredentialName: value });
  }

  onSelectGitCredential(selectedGitCredential) {
    return this.$async(async () => {
      if (selectedGitCredential) {
        this.formValues.SelectedGitCredential = selectedGitCredential;
        this.formValues.RepositoryGitCredentialID = selectedGitCredential.id;
        this.formValues.RepositoryUsername = selectedGitCredential.username;
        this.formValues.SaveGitCredential = false;
        this.formValues.NewCredentialName = '';
      } else {
        this.formValues.SelectedGitCredential = null;
        this.formValues.RepositoryUsername = '';
        this.formValues.RepositoryGitCredentialID = 0;
      }

      this.formValues.RepositoryPassword = '';

      this.state.hasUnsavedChanges = angular.toJson(this.savedFormValues) !== angular.toJson(this.formValues);
    });
  }

  onChangeAutoUpdate(values) {
    this.onChange({
      AutoUpdate: {
        ...this.formValues.AutoUpdate,
        ...values,
      },
    });
  }

  onChangeEnvVar(value) {
    this.onChange({ Env: value });
  }

  onChangeOption(values) {
    this.onChange({
      Option: {
        ...this.formValues.Option,
        ...values,
      },
    });
  }

  async submit() {
    const isSwarmStack = this.stack.Type === 1;
    const that = this;
    this.ModalService.confirmStackUpdate(
      'Any changes to this stack or application made locally in Portainer will be overridden, which may cause service interruption. Do you wish to continue',
      isSwarmStack,
      'btn-warning',
      async function (result) {
        if (!result) {
          return;
        }
        try {
          that.state.redeployInProgress = true;

          // save git credential
          if (that.formValues.SaveCredential && that.formValues.NewCredentialName) {
            const userDetails = this.Authentication.getUserDetails();
            await that.UserService.saveGitCredential(
              userDetails.ID,
              that.formValues.NewCredentialName,
              that.formValues.RepositoryUsername,
              that.formValues.RepositoryPassword
            ).then(function success(data) {
              that.formValues.RepositoryGitCredentialID = data.gitCredential.id;
            });
          }

          await that.StackService.updateGit(
            that.stack.Id,
            that.stack.EndpointId,
            that.FormHelper.removeInvalidEnvVars(that.formValues.Env),
            that.formValues.Option.Prune,
            that.formValues,
            !!result[0]
          );
          that.Notifications.success('Success', 'Pulled and redeployed stack successfully');
          that.$state.reload();
        } catch (err) {
          that.Notifications.error('Failure', err, 'Failed redeploying stack');
        } finally {
          that.state.redeployInProgress = false;
        }
      }
    );
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
              that.formValues.RepositoryGitCredentialID = data.gitCredential.id;
            }
          );
        }

        const stack = await this.StackService.updateGitStackSettings(
          this.stack.Id,
          this.stack.EndpointId,
          this.FormHelper.removeInvalidEnvVars(this.formValues.Env),
          this.formValues
        );
        this.savedFormValues = angular.copy(this.formValues);
        this.state.hasUnsavedChanges = false;
        this.Notifications.success('Success', 'Save stack settings successfully');

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

  isSubmitButtonDisabled() {
    return this.state.inProgress || this.state.redeployInProgress;
  }

  isAutoUpdateChanged() {
    const wasEnabled = !!(this.stack.AutoUpdate && (this.stack.AutoUpdate.Interval || this.stack.AutoUpdate.Webhook));
    const isEnabled = this.formValues.AutoUpdate.RepositoryAutomaticUpdates;
    return isEnabled !== wasEnabled;
  }

  async $onInit() {
    this.formValues.RefName = this.model.ReferenceName;
    this.formValues.Env = this.stack.Env;

    try {
      this.formValues.GitCredentials = await this.UserService.getGitCredentials(this.Authentication.getUserDetails().ID);
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve user saved git credentials');
    }

    if (this.stack.Option) {
      this.formValues.Option = this.stack.Option;
    }

    // Init auto update
    if (this.stack.AutoUpdate && (this.stack.AutoUpdate.Interval || this.stack.AutoUpdate.Webhook)) {
      this.formValues.AutoUpdate.RepositoryAutomaticUpdates = true;
      this.formValues.AutoUpdate.RepositoryAutomaticUpdatesForce = this.stack.AutoUpdate.ForceUpdate;
      this.formValues.AutoUpdate.ForcePullImage = this.stack.AutoUpdate.ForcePullImage;

      if (this.stack.AutoUpdate.Interval) {
        this.formValues.AutoUpdate.RepositoryMechanism = RepositoryMechanismTypes.INTERVAL;
        this.formValues.AutoUpdate.RepositoryFetchInterval = this.stack.AutoUpdate.Interval;
      } else if (this.stack.AutoUpdate.Webhook) {
        this.formValues.AutoUpdate.RepositoryMechanism = RepositoryMechanismTypes.WEBHOOK;
        this.formValues.AutoUpdate.RepositoryWebhookURL = this.WebhookHelper.returnStackWebhookUrl(this.stack.AutoUpdate.Webhook);
      }
    }

    if (!this.formValues.AutoUpdate.RepositoryWebhookURL) {
      this.formValues.AutoUpdate.RepositoryWebhookURL = this.WebhookHelper.returnStackWebhookUrl(uuidv4());
    }

    if (this.stack.GitConfig && this.stack.GitConfig.Authentication) {
      this.formValues.RepositoryUsername = this.stack.GitConfig.Authentication.Username;
      this.formValues.RepositoryAuthentication = true;
      this.state.isEdit = true;

      if (this.stack.GitConfig.Authentication.GitCredentialID > 0) {
        this.formValues.SelectedGitCredential = this.formValues.GitCredentials.find((x) => x.id === this.stack.GitConfig.Authentication.GitCredentialID);
        if (this.formValues.SelectedGitCredential) {
          this.formValues.RepositoryGitCredentialID = this.formValues.SelectedGitCredential.id;
          this.formValues.RepositoryUsername = this.formValues.SelectedGitCredential.username;
          this.formValues.RepositoryPassword = '';
        }
      }
    }
    this.savedFormValues = angular.copy(this.formValues);
  }
}

export default StackRedeployGitFormController;
