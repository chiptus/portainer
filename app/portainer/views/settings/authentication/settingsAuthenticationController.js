import angular from 'angular';
import _ from 'lodash-es';

import { buildLdapSettingsModel, buildAdSettingsModel } from '@/portainer/settings/authentication/ldap/ldap-settings.model';
import { options } from '@/react/portainer/settings/AuthenticationView/InternalAuth/options';

angular.module('portainer.app').controller('SettingsAuthenticationController', SettingsAuthenticationController);

function SettingsAuthenticationController($q, $scope, $state, Notifications, SettingsService, FileUploadService, TeamService, LDAPService) {
  $scope.authMethod = 1;

  $scope.state = {
    uploadInProgress: false,
    actionInProgress: false,
    availableUserSessionTimeoutOptions: [
      {
        key: '1 hour',
        value: '1h',
      },
      {
        key: '4 hours',
        value: '4h',
      },
      {
        key: '8 hours',
        value: '8h',
      },
      {
        key: '24 hours',
        value: '24h',
      },
      { key: '1 week', value: `${24 * 7}h` },
      { key: '1 month', value: `${24 * 30}h` },
      { key: '6 months', value: `${24 * 30 * 6}h` },
      { key: '1 year', value: `${24 * 30 * 12}h` },
    ],
  };

  $scope.formValues = {
    UserSessionTimeout: $scope.state.availableUserSessionTimeoutOptions[0],
    TLSCACert: '',
    ldap: {
      serverType: 0,
      adSettings: buildAdSettingsModel(),
      ldapSettings: buildLdapSettingsModel(),
    },
    selectedAdminGroups: [],
  };

  $scope.authOptions = options;

  $scope.onChangeAuthMethod = function onChangeAuthMethod(value) {
    $scope.authMethod = value;

    if (value === 4) {
      $scope.settings.AuthenticationMethod = 2;
      $scope.formValues.ldap.serverType = 2;
      return;
    }

    if (value === 2) {
      $scope.settings.AuthenticationMethod = 2;
      $scope.formValues.ldap.serverType = $scope.formValues.ldap.ldapSettings.ServerType;
      return;
    }

    $scope.settings.AuthenticationMethod = value;
  };

  $scope.onChangePasswordLength = function onChangePasswordLength(value) {
    $scope.$evalAsync(() => {
      $scope.settings.InternalAuthSettings = { RequiredPasswordLength: value };
    });
  };

  $scope.authenticationMethodSelected = function authenticationMethodSelected(value) {
    if (!$scope.settings) {
      return false;
    }

    if (value === 4) {
      return $scope.settings.AuthenticationMethod === 2 && $scope.formValues.ldap.serverType === 2;
    }

    if (value === 2) {
      return $scope.settings.AuthenticationMethod === 2 && $scope.formValues.ldap.serverType !== 2;
    }

    return $scope.settings.AuthenticationMethod === value;
  };

  $scope.isOauthEnabled = function isOauthEnabled() {
    return $scope.settings && $scope.settings.AuthenticationMethod === 3;
  };

  $scope.LDAPConnectivityCheck = LDAPConnectivityCheck;
  function LDAPConnectivityCheck() {
    const settings = angular.copy($scope.settings);

    const { settings: ldapSettings, uploadRequired, tlscaFile } = prepareLDAPSettings();
    settings.LDAPSettings = ldapSettings;
    $scope.state.uploadInProgress = uploadRequired;

    $scope.state.connectivityCheckInProgress = true;

    $q.when(!uploadRequired || FileUploadService.uploadLDAPTLSFiles(tlscaFile, null, null))
      .then(function success() {
        return LDAPService.check(settings.LDAPSettings);
      })
      .then(function success() {
        $scope.state.failedConnectivityCheck = false;
        $scope.state.successfulConnectivityCheck = true;
        Notifications.success('Success', 'Connection to LDAP successful');
      })
      .catch(function error(err) {
        $scope.state.failedConnectivityCheck = true;
        $scope.state.successfulConnectivityCheck = false;
        Notifications.error('Failure', err, 'Connection to LDAP failed');
      })
      .finally(function final() {
        $scope.state.uploadInProgress = false;
        $scope.state.connectivityCheckInProgress = false;
      });
  }

  $scope.saveSettings = function () {
    const settings = angular.copy($scope.settings);

    if (!settings.OAuthSettings.AdminAutoPopulate) {
      delete settings.OAuthSettings.AdminGroupClaimsRegexList;
    }

    const { settings: ldapSettings, uploadRequired, tlscaFile } = prepareLDAPSettings();
    settings.LDAPSettings = ldapSettings;
    $scope.state.uploadInProgress = uploadRequired;

    $scope.state.actionInProgress = true;

    $q.when(!uploadRequired || FileUploadService.uploadLDAPTLSFiles(tlscaFile, null, null))
      .then(function success() {
        return SettingsService.update(settings);
      })
      .then(function success() {
        Notifications.success('Success', 'Authentication settings updated');
      })
      .catch(function error(err) {
        Notifications.error('Failure', err, 'Unable to update authentication settings');
      })
      .finally(function final() {
        $scope.state.uploadInProgress = false;
        $scope.state.actionInProgress = false;
      });
  };

  function prepareLDAPSettings() {
    const tlscaCert = $scope.formValues.TLSCACert;

    const tlscaFile = tlscaCert !== $scope.settings.LDAPSettings.TLSConfig.TLSCACert ? tlscaCert : null;

    const isADServer = $scope.formValues.ldap.serverType === 2;

    const settings = isADServer ? $scope.formValues.ldap.adSettings : $scope.formValues.ldap.ldapSettings;

    if (settings.AnonymousMode && !isADServer) {
      settings.ReaderDN = '';
      settings.Password = '';
    }

    if (isADServer) {
      settings.AnonymousMode = false;
    }

    settings.URLs = settings.URLs.map((url) => {
      if (url === undefined || url === '') {
        return;
      }

      if (url.includes(':')) {
        return url;
      }
      return url + (settings.TLSConfig.TLS ? ':636' : ':389');
    });

    const uploadRequired = (settings.TLSConfig.TLS || settings.StartTLS) && !settings.TLSConfig.TLSSkipVerify;

    if (settings.AdminAutoPopulate && $scope.formValues.selectedAdminGroups.length > 0) {
      settings.AdminGroups = $scope.formValues.selectedAdminGroups;
    } else {
      settings.AdminGroups = [];
    }

    if ($scope.formValues.selectedAdminGroups.length === 0) {
      settings.AdminAutoPopulate = false;
    }

    return { settings, uploadRequired, tlscaFile };
  }

  $scope.isLDAPFormValid = isLDAPFormValid;
  function isLDAPFormValid() {
    const ldapSettings = $scope.formValues.ldap.serverType === 2 ? $scope.formValues.ldap.adSettings : $scope.formValues.ldap.ldapSettings;
    const isTLSMode = ldapSettings.TLSConfig.TLS || ldapSettings.StartTLS;

    return (
      _.compact(ldapSettings.URLs).length &&
      (ldapSettings.AnonymousMode || (ldapSettings.ReaderDN && ldapSettings.Password)) &&
      (!isTLSMode || (isTLSMode && $scope.formValues.TLSCACert) || ldapSettings.TLSConfig.TLSSkipVerify) &&
      (!$scope.settings.LDAPSettings.AdminAutoPopulate || ($scope.settings.LDAPSettings.AdminAutoPopulate && $scope.formValues.selectedAdminGroups.length > 0))
    );
  }

  $scope.isOAuthTeamMembershipFormValid = isOAuthTeamMembershipFormValid;
  function isOAuthTeamMembershipFormValid() {
    if ($scope.settings && $scope.settings.OAuthSettings.OAuthAutoMapTeamMemberships && $scope.settings.OAuthSettings.TeamMemberships) {
      if (!$scope.settings.OAuthSettings.TeamMemberships.OAuthClaimName) {
        return false;
      }

      const hasInvalidMapping = $scope.settings.OAuthSettings.TeamMemberships.OAuthClaimMappings.some((m) => !(m.ClaimValRegex && m.Team));
      if (hasInvalidMapping) {
        return false;
      }
    }
    return true;
  }

  $scope.isOAuthAdminMappingFormValid = isOAuthAdminMappingFormValid;
  function isOAuthAdminMappingFormValid() {
    if ($scope.settings && $scope.settings.OAuthSettings.AdminAutoPopulate && $scope.settings.OAuthSettings.AdminGroupClaimsRegexList) {
      const hasInvalidMapping =
        $scope.settings.OAuthSettings.AdminGroupClaimsRegexList.length === 0 || $scope.settings.OAuthSettings.AdminGroupClaimsRegexList.some((e) => e === '');
      if (hasInvalidMapping) {
        return false;
      }
    }

    return true;
  }

  function initView() {
    $q.all({
      settings: SettingsService.settings(),
      teams: TeamService.teams(),
    })
      .then(function success(data) {
        var settings = data.settings;
        $scope.teams = data.teams;
        $scope.settings = settings;

        $scope.OAuthSettings = settings.OAuthSettings;
        $scope.authMethod = settings.AuthenticationMethod;
        if (settings.AuthenticationMethod === 2 && settings.LDAPSettings.ServerType === 2) {
          $scope.authMethod = 4;
        }

        $scope.formValues.ldap.serverType = settings.LDAPSettings.ServerType;
        if (settings.LDAPSettings.ServerType === 2) {
          $scope.formValues.ldap.adSettings = settings.LDAPSettings;
        } else {
          $scope.formValues.ldap.ldapSettings = settings.LDAPSettings;
        }

        $scope.formValues.selectedAdminGroups = settings.LDAPSettings.AdminGroups || [];

        if (!settings.LDAPSettings.URLs.length) {
          settings.LDAPSettings.URLs.push('');
        }
      })
      .catch(function error(err) {
        Notifications.error('Failure', err, 'Unable to retrieve application settings');
      });
  }

  initView();
}
