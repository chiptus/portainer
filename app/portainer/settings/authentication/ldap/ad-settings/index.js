import controller from './ad-settings.controller';

export const adSettings = {
  templateUrl: './ad-settings.html',
  controller,
  bindings: {
    settings: '=',
    selectedAdminGroups: '=',
    tlscaCert: '=',
    state: '=',
    connectivityCheck: '<',
  },
};
