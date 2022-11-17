import _ from 'lodash';
import axios from '@/portainer/services/axios';

export default class GitFormController {
  /* @ngInject */
  constructor($scope, Notifications, StateManager) {
    Object.assign(this, { $scope, Notifications, StateManager });

    this.onChangeField = this.onChangeField.bind(this);
    this.onChangeURL = this.onChangeField('RepositoryURL');
    this.onChangeRepositoryValid = this.onChangeRepositoryValid.bind(this);
    this.onChangeRefName = this.onChangeField('RepositoryReferenceName');
    this.onChangeComposePath = this.onChangeField('ComposeFilePathInRepository');
    this.refreshGitopsCache = this.refreshGitopsCache.bind(this);
    this.onRefreshGitopsCache = this.onRefreshGitopsCache.bind(this);
    this.$scope.$watch(() => this.model.SelectedGitCredential, _.debounce(this.refreshGitopsCache, 300));
    this.$scope.$watch(() => this.model.RepositoryUsername, _.debounce(this.refreshGitopsCache, 300));
    this.$scope.$watch(() => this.model.RepositoryPassword, _.debounce(this.refreshGitopsCache, 300));
    this.onChangeRepositoryUsername = this.onChangeField('RepositoryUsername');
    this.onChangeRepositoryPassword = this.onChangeField('RepositoryPassword');
    this.onChangeSaveCredential = this.onChangeField('SaveCredential');
    this.onChangeNewCredentialName = this.onChangeField('NewCredentialName');
    this.onChangeRepositoryAuthentication = this.onChangeField('RepositoryAuthentication');
  }

  handleChange(...args) {
    this.$scope.$evalAsync(() => {
      this.onChange(...args);
    });
  }

  onChangeField(field) {
    return (value) => {
      this.handleChange({
        ...this.model,
        [field]: value,
      });
    };
  }

  onChangeRepositoryValid(valid) {
    this.handleChange({
      ...this.model,
      RepositoryURLValid: valid,
    });
  }

  refreshGitopsCache() {
    if (
      (this.model.RepositoryAuthentication && !this.model.RepositoryPassword && !this.model.SelectedGitCredential) ||
      !this.model.RepositoryURL ||
      this.model.RepositoryURLValid === false
    ) {
      return;
    }

    const payload = {
      repository: this.model.RepositoryURL,
    };

    if (this.model.SelectedGitCredential) {
      payload.gitCredentialId = this.model.SelectedGitCredential.id;
    } else {
      payload.username = this.model.RepositoryUsername;
      payload.password = this.model.RepositoryPassword;
    }
    axios.post('/gitops/repo/refs', payload, {
      params: { force: true },
    });
  }

  onRefreshGitopsCache() {
    this.refreshGitopsCache();
  }

  $onInit() {
    this.deployMethod = this.deployMethod || 'compose';
    this.isDockerStandalone = !this.hideRebuildInfo && this.StateManager.getState().endpoint.mode.provider === 'DOCKER_STANDALONE';
  }
}
