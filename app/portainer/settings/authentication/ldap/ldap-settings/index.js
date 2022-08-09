import controller from './ldap-settings.controller';

export const ldapSettings = {
  templateUrl: './ldap-settings.html',
  controller,
  bindings: {
    settings: '=',
    tlscaCert: '=',
    selectedAdminGroups: '=',
    state: '<',
    connectivityCheck: '<',
    onSaveSettings: '<',
    saveButtonState: '<',
    isLdapFormValid: '<',
  },
};
