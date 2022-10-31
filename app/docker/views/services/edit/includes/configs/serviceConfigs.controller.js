import { notifyError } from '@/portainer/services/notifications';

/* @ngInject */
export default function ServiceConfigsController($async, StateManager, ConfigService) {
  this.applicationState = StateManager.getState();

  this.configs = [];

  this.addConfig = function addConfig(config) {
    if (
      config &&
      this.value.every(function (serviceConfig) {
        return serviceConfig.Id !== config.Id;
      })
    ) {
      this.value.push({ Id: config.Id, Name: config.Name, FileName: config.Name, Uid: '0', Gid: '0', Mode: 292 });
      this.onChange(this.value);
    }
  };

  this.removeConfig = function removeSecret(index) {
    var removedElement = this.value.splice(index, 1);
    if (removedElement !== null) {
      this.onChange(this.value);
    }
  };

  this.updateConfig = function updateConfig() {
    this.onChange(this.value);
  };

  // this.handleUpdate = function handleUpdate(configs) {
  //   this.this.value = configs;
  //   this.updateServiceArray(this.service, 'ServiceConfigs',configs);
  // }

  this.$onInit = function $onInit() {
    return $async(async () => {
      var apiVersion = this.applicationState.endpoint.apiVersion;

      try {
        this.configs = apiVersion >= 1.3 ? await ConfigService.configs() : [];
      } catch (err) {
        notifyError('Failure', err, 'Unable to retrieve service configs');
      }
    });
  };
}
