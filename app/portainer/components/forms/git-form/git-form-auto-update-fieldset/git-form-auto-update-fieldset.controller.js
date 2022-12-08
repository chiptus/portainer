import { FeatureId } from '@/react/portainer/feature-flags/enums';

class GitFormAutoUpdateFieldsetController {
  /* @ngInject */
  constructor($scope, clipboard, StateManager) {
    Object.assign(this, { $scope, clipboard, StateManager });

    this.onChangeAutoUpdate = this.onChangeField('RepositoryAutomaticUpdates');
    this.onChangeAutoUpdateForce = this.onChangeField('RepositoryAutomaticUpdatesForce');
    this.onChangeMechanism = this.onChangeField('RepositoryMechanism');
    this.onChangeInterval = this.onChangeField('RepositoryFetchInterval');
    this.onChangeForcePullImage = this.onChangeField('ForcePullImage');

    this.limitedFeature = FeatureId.FORCE_REDEPLOYMENT;
    this.limitedFeaturePullImage = FeatureId.STACK_PULL_IMAGE;
  }

  copyWebhook() {
    this.clipboard.copyText(this.model.RepositoryWebhookURL);
    $('#copyNotification').show();
    $('#copyNotification').fadeOut(2000);
  }

  onChangeField(field) {
    return (value) => {
      this.$scope.$evalAsync(() => {
        this.onChange({
          ...this.model,
          [field]: value,
        });
      });
    };
  }

  $onInit() {
    this.environmentType = this.StateManager.getState().endpoint.mode.provider;
  }
}

export default GitFormAutoUpdateFieldsetController;
