import uuidv4 from 'uuid/v4';
import { RepositoryMechanismTypes } from 'Kubernetes/models/deploy';
class StackRedeployGitFormController {
  /* @ngInject */
  constructor($async, $state, StackService, ModalService, Notifications, WebhookHelper, FormHelper) {
    this.$async = $async;
    this.$state = $state;
    this.StackService = StackService;
    this.ModalService = ModalService;
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
      Env: [],
      PullImage: false,
      // auto update
      AutoUpdate: {
        RepositoryAutomaticUpdates: false,
        RepositoryAutomaticUpdatesForce: false,
        RepositoryMechanism: RepositoryMechanismTypes.INTERVAL,
        RepositoryFetchInterval: '5m',
        RepositoryWebhookURL: '',
        ForcePullImage: false,
        ShowForcePullImage: false,
      },
    };

    this.onChange = this.onChange.bind(this);
    this.onChangeRef = this.onChangeRef.bind(this);
    this.onChangeAutoUpdate = this.onChangeAutoUpdate.bind(this);
    this.onChangeEnvVar = this.onChangeEnvVar.bind(this);
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
    this.formValues = {
      ...this.formValues,
      ...values,
    };

    this.state.hasUnsavedChanges = angular.toJson(this.savedFormValues) !== angular.toJson(this.formValues);
  }

  onChangeRef(value) {
    this.onChange({ RefName: value });
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

  async submit() {
    const isSwarmStack = this.stack.Type === 1;
    const that = this;
    this.ModalService.confirmStackUpdate(
      'Any changes to this stack or application made locally in Portainer will be overridden, which may cause service interruption. Do you wish to continue',
      isSwarmStack,
      'btn-warning',
      function (result) {
        if (!result) {
          return;
        }
        try {
          that.state.redeployInProgress = true;
          that.StackService.updateGit(that.stack.Id, that.stack.EndpointId, that.FormHelper.removeInvalidEnvVars(that.formValues.Env), false, that.formValues, !!result[0]);
          that.Notifications.success('Pulled and redeployed stack successfully');
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
    return this.$async(async () => {
      try {
        this.state.inProgress = true;
        const stack = await this.StackService.updateGitStackSettings(
          this.stack.Id,
          this.stack.EndpointId,
          this.FormHelper.removeInvalidEnvVars(this.formValues.Env),
          this.formValues
        );
        this.savedFormValues = angular.copy(this.formValues);
        this.state.hasUnsavedChanges = false;
        this.Notifications.success('Save stack settings successfully');

        this.stack = stack;
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to save stack settings');
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

  $onInit() {
    this.formValues.RefName = this.model.ReferenceName;
    this.formValues.Env = this.stack.Env;
    // Init auto update
    if (this.stack.AutoUpdate && (this.stack.AutoUpdate.Interval || this.stack.AutoUpdate.Webhook)) {
      this.formValues.AutoUpdate.RepositoryAutomaticUpdates = true;
      this.formValues.AutoUpdate.RepositoryAutomaticUpdatesForce = this.stack.AutoUpdate.ForceUpdate;
      this.formValues.AutoUpdate.ForcePullImage = this.stack.AutoUpdate.ForcePullImage;
      this.formValues.AutoUpdate.ShowForcePullImage = this.stack.Type !== 3;

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
    }

    this.savedFormValues = angular.copy(this.formValues);
  }
}

export default StackRedeployGitFormController;
