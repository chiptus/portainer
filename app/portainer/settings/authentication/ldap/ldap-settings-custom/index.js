import controller from './ldap-settings-custom.controller';

export const ldapSettingsCustom = {
  templateUrl: './ldap-settings-custom.html',
  controller,
  bindings: {
    settings: '=',
    tlscaCert: '=',
    state: '=',
    selectedAdminGroups: '=',
    onTlscaCertChange: '<',
    connectivityCheck: '<',
    onSearchUsersClick: '<',
    onSearchGroupsClick: '<',
    onSearchAdminGroupsClick: '<',
    onSaveSettings: '<',
    saveButtonState: '<',
    saveButtonDisabled: '<',
  },
};
