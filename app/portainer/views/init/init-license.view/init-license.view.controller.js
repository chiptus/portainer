export default class InitLicenseViewController {
  /* @ngInject */
  constructor($async, $state, $scope, $location, LicenseService, Notifications, Authentication) {
    this.$async = $async;
    this.$state = $state;
    this.$scope = $scope;
    this.$location = $location;
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

  $onInit() {
    // Prevent bypassing license init view
    this.$scope.$on('$locationChangeStart', () => {
      this.$location.path('/init/license');
    });
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
        await this.LicenseService.attach({ key: this.license }, { params: { force: true } });

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
