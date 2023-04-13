class SslCaFileSettingsController {
  /* @ngInject */
  constructor($async, $scope, $state, SSLService, Notifications) {
    Object.assign(this, { $async, $scope, $state, SSLService, Notifications });

    this.clientCert = null;
    this.originalValues = {
      clientCert: null,
    };

    this.formValues = {
      clientCert: null,
    };

    this.state = {
      actionInProgress: false,
      reloadingPage: false,
    };

    this.certFilePattern = '.pem,.crt,.cer,.cert';

    this.save = this.save.bind(this);
  }

  isFormChanged() {
    return Object.entries(this.originalValues).some(([key, value]) => value != this.formValues[key]);
  }

  async save() {
    this.state.actionInProgress = true;
    try {
      const clientCert = this.formValues.clientCert ? await this.formValues.clientCert.text() : null;

      await this.SSLService.upload({ clientCert });
      await new Promise((resolve) => setTimeout(resolve, 2000));
      location.reload();
      this.state.reloadingPage = true;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Failed applying changes');
    }
    this.state.actionInProgress = false;
  }
}

export default SslCaFileSettingsController;
