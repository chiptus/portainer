export default class InitLicenseViewController {
  /* @ngInject */
  constructor($async, $state, LicenseService, Notifications, Authentication) {
    this.$async = $async;
    this.$state = $state;
    this.LicenseService = LicenseService;
    this.Notifications = Notifications;
    this.Authentication = Authentication;

    this.license = '';
    this.state = {
      actionInProgress: false,
      formError: '',
    };

    this.submit = this.submit.bind(this);
  }

  isFormValid() {
    return !!this.license;
  }

  async submit() {
    return this.$async(async () => {
      this.state.formError = '';
      if (!this.isFormValid()) {
        this.state.formError = 'Form is invalid';
        return;
      }

      this.state.actionInProgress = true;

      try {
        const { failedKeys } = await this.LicenseService.attach([this.license]);

        if (failedKeys[this.license]) {
          throw new Error(failedKeys[this.license]);
        }

        let path = 'portainer.wizard';
        if (this.Authentication.getUserDetails().forceChangePassword) {
          path = 'portainer.account';
        }
        this.$state.go(path);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Failed validating licenses');
      } finally {
        this.state.actionInProgress = false;
      }
    });
  }
}
