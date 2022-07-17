import { RegistryTypes } from 'Portainer/models/registryTypes';
import { RegistryManagementConfigurationDefaultModel } from 'Portainer/models/registry';

export class ConfigureRegistryController {
  /* @ngInject */
  constructor($scope, $async, $state, RegistryService, RegistryServiceSelector, Notifications) {
    this.$scope = $scope;
    Object.assign(this, { $async, $state, RegistryService, RegistryServiceSelector, Notifications });

    this.state = {
      testInProgress: false,
      updateInProgress: false,
      validConfiguration: false,
    };

    this.testConfiguration = this.testConfiguration.bind(this);
    this.testConfigurationAsync = this.testConfigurationAsync.bind(this);
    this.updateConfiguration = this.updateConfiguration.bind(this);
    this.updateConfigurationAsync = this.updateConfigurationAsync.bind(this);
    this.toggleAuthentication = this.toggleAuthentication.bind(this);
    this.toggleTLS = this.toggleTLS.bind(this);
    this.toggleTLSSkipVerify = this.toggleTLSSkipVerify.bind(this);
    this.$onInit = this.$onInit.bind(this);
  }

  toggleAuthentication(newValue) {
    this.$scope.$evalAsync(() => {
      this.model.Authentication = newValue;
    });
  }

  toggleTLS(newValue) {
    this.$scope.$evalAsync(() => {
      this.model.TLS = newValue;
    });
  }

  toggleTLSSkipVerify(newValue) {
    this.$scope.$evalAsync(() => {
      this.model.TLSSkipVerify = newValue;
    });
  }

  testConfiguration() {
    return this.$async(this.testConfigurationAsync);
  }
  async testConfigurationAsync() {
    this.state.testInProgress = true;
    try {
      await this.RegistryService.configureRegistry(this.registry.Id, this.model);
      await this.RegistryServiceSelector.ping(this.registry, null, true);

      this.Notifications.success('Success', 'Valid management configuration');
      this.state.validConfiguration = true;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Invalid management configuration');
    }

    this.state.testInProgress = false;
  }

  updateConfiguration() {
    return this.$async(this.updateConfigurationAsync);
  }
  async updateConfigurationAsync() {
    this.state.updateInProgress = true;
    try {
      await this.RegistryService.configureRegistry(this.registry.Id, this.model);
      this.Notifications.success('Success', 'Registry management configuration updated');
      this.$state.go('portainer.registries.registry.repositories', { id: this.registry.Id }, { reload: true });
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to update registry management configuration');
    }

    this.state.updateInProgress = false;
  }

  async $onInit() {
    const registryId = this.$state.params.id;
    this.RegistryTypes = RegistryTypes;

    try {
      const registry = await this.RegistryService.registry(registryId);
      const model = new RegistryManagementConfigurationDefaultModel(registry);

      this.registry = registry;
      this.model = model;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve registry details');
    }
  }
}
