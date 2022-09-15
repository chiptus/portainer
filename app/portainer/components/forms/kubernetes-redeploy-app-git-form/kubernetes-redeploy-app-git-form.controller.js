import uuidv4 from 'uuid/v4';
import { RepositoryMechanismTypes } from 'Kubernetes/models/deploy';
class KubernetesRedeployAppGitFormController {
  /* @ngInject */
  constructor($async, $state, $analytics, StackService, ModalService, UserService, Authentication, Notifications, WebhookHelper) {
    this.$async = $async;
    this.$state = $state;
    this.$analytics = $analytics;
    this.StackService = StackService;
    this.ModalService = ModalService;
    this.UserService = UserService;
    this.Authentication = Authentication;
    this.Notifications = Notifications;
    this.WebhookHelper = WebhookHelper;

    this.state = {
      saveGitSettingsInProgress: false,
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
      // auto update
      AutoUpdate: {
        RepositoryAutomaticUpdates: false,
        RepositoryAutomaticUpdatesForce: false,
        RepositoryMechanism: RepositoryMechanismTypes.INTERVAL,
        RepositoryFetchInterval: '5m',
        RepositoryWebhookURL: '',
      },
    };

    this.onChange = this.onChange.bind(this);
    this.onChangeRef = this.onChangeRef.bind(this);
    this.onChangeAutoUpdate = this.onChangeAutoUpdate.bind(this);
    this.onSelectGitCredential = this.onSelectGitCredential.bind(this);
    this.onChangeSaveCredential = this.onChangeSaveCredential.bind(this);
    this.onChangeNewCredentialName = this.onChangeNewCredentialName.bind(this);
    this.onChangeRepositoryAuthentication = this.onChangeRepositoryAuthentication.bind(this);
    this.onChangeRepositoryUsername = this.onChangeRepositoryUsername.bind(this);
    this.onChangeRepositoryPassword = this.onChangeRepositoryPassword.bind(this);
    this.onChangeSaveCredential = this.onChangeSaveCredential.bind(this);
  }

  onChangeRef(value) {
    this.onChange({ RefName: value });
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

  buildAnalyticsProperties() {
    const metadata = {
      'automatic-updates': automaticUpdatesLabel(this.formValues.AutoUpdate.RepositoryAutomaticUpdates, this.formValues.AutoUpdate.RepositoryMechanism),
    };

    return { metadata };

    function automaticUpdatesLabel(repositoryAutomaticUpdates, repositoryMechanism) {
      switch (repositoryAutomaticUpdates && repositoryMechanism) {
        case RepositoryMechanismTypes.INTERVAL:
          return 'polling';
        case RepositoryMechanismTypes.WEBHOOK:
          return 'webhook';
        default:
          return 'off';
      }
    }
  }

  async pullAndRedeployApplication() {
    const that = this;
    return this.$async(async () => {
      try {
        const confirmed = await this.ModalService.confirmAsync({
          title: 'Are you sure?',
          message: 'Any changes to this application will be overridden by the definition in git and may cause a service interruption. Do you wish to continue?',
          buttons: {
            confirm: {
              label: 'Update',
              className: 'btn-warning',
            },
          },
        });
        if (!confirmed) {
          return;
        }
        this.state.redeployInProgress = true;
        // save git credential
        if (that.formValues.SaveCredential && that.formValues.NewCredentialName) {
          const userDetails = this.Authentication.getUserDetails();
          await that.UserService.saveGitCredential(userDetails.ID, that.formValues.NewCredentialName, that.formValues.RepositoryUsername, that.formValues.RepositoryPassword).then(
            function success(data) {
              that.formValues.RepositoryGitCredentialID = data.gitCredential.id;
            }
          );
        }
        await this.StackService.updateKubeGit(this.stack.Id, this.stack.EndpointId, this.namespace, this.formValues);
        this.Notifications.success('Success', 'Pulled and redeployed stack successfully');
        await this.$state.reload();
      } catch (err) {
        this.Notifications.error('Failure', err, 'Failed redeploying application');
      } finally {
        this.state.redeployInProgress = false;
      }
    });
  }

  async saveGitSettings() {
    const that = this;
    const userDetails = this.Authentication.getUserDetails();
    return this.$async(async () => {
      try {
        this.state.saveGitSettingsInProgress = true;

        this.state.inProgress = true;

        // save git credential
        if (this.formValues.SaveCredential && this.formValues.NewCredentialName) {
          await this.UserService.saveGitCredential(userDetails.ID, this.formValues.NewCredentialName, this.formValues.RepositoryUsername, this.formValues.RepositoryPassword).then(
            function success(data) {
              that.formValues.RepositoryGitCredentialID = data.gitCredential.id;
            }
          );
        }

        await this.StackService.updateKubeStack({ EndpointId: this.stack.EndpointId, Id: this.stack.Id }, null, this.formValues);
        this.savedFormValues = angular.copy(this.formValues);
        this.state.hasUnsavedChanges = false;
        this.Notifications.success('Success', 'Save stack settings successfully');
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to save application settings');
        if (that.formValues.SaveCredential && that.formValues.NewCredentialName && that.formValues.RepositoryGitCredentialID) {
          that.UserService.deleteGitCredential(userDetails.ID, that.formValues.RepositoryGitCredentialID);
        }
      } finally {
        this.state.saveGitSettingsInProgress = false;
      }
    });
  }

  isSubmitButtonDisabled() {
    return this.state.saveGitSettingsInProgress || this.state.redeployInProgress;
  }

  async $onInit() {
    this.formValues.RefName = this.stack.GitConfig.ReferenceName;

    try {
      this.formValues.GitCredentials = await this.UserService.getGitCredentials(this.Authentication.getUserDetails().ID);
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve user saved git credentials');
    }

    // Init auto update
    if (this.stack.AutoUpdate && (this.stack.AutoUpdate.Interval || this.stack.AutoUpdate.Webhook)) {
      this.formValues.AutoUpdate.RepositoryAutomaticUpdates = true;

      this.formValues.AutoUpdate.RepositoryAutomaticUpdatesForce = this.stack.AutoUpdate.ForceUpdate;

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

export default KubernetesRedeployAppGitFormController;
