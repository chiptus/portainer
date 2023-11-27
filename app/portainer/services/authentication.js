import { getCurrentUser } from '../users/queries/useLoadCurrentUser';
import { isAdmin as isAdminHelperFunc, isPureAdmin as isPureAdminHelperFunc } from '../users/user.helpers';
import { clear as clearSessionStorage } from './session-storage';
angular.module('portainer.app').factory('Authentication', [
  '$async',
  '$state',
  'Auth',
  'OAuth',
  'LocalStorage',
  'StateManager',
  'EndpointProvider',
  'ThemeManager',
  function AuthenticationFactory($async, $state, Auth, OAuth, LocalStorage, StateManager, EndpointProvider, ThemeManager) {
    'use strict';

    var user = {};

    if (process.env.NODE_ENV === 'development') {
      window.login = loginAsync;
    }

    return {
      init,
      OAuthLogin,
      login,
      logout,
      isAuthenticated,
      getUserDetails,
      isAdmin,
      isPureAdmin,
      hasAuthorizations,
      redirectIfUnauthorized,
    };

    async function initAsync() {
      try {
        const userId = LocalStorage.getUserId();
        if (userId && user.ID === userId) {
          return true;
        }
        await loadUserData();
        return true;
      } catch (error) {
        return false;
      }
    }

    async function logoutAsync() {
      if (isAuthenticated()) {
        await Auth.logout().$promise;
      }

      clearSessionStorage();
      StateManager.clean();
      EndpointProvider.clean();
      LocalStorage.cleanAuthData();
      LocalStorage.storeLoginStateUUID('');
      cleanUserData();
    }

    function logout() {
      return $async(logoutAsync);
    }

    function init() {
      return $async(initAsync);
    }

    async function OAuthLoginAsync(code) {
      await OAuth.validate({ code: code }).$promise;
      await loadUserData();
    }

    function OAuthLogin(code) {
      return $async(OAuthLoginAsync, code);
    }

    async function loginAsync(username, password) {
      await Auth.login({ username: username, password: password }).$promise;
      await loadUserData();
    }

    function login(username, password) {
      return $async(loginAsync, username, password);
    }

    function isAuthenticated() {
      return !!user.ID;
    }

    function getUserDetails() {
      return user;
    }

    function cleanUserData() {
      user = {};
    }

    async function loadUserData() {
      const userData = await getCurrentUser();
      user.username = userData.Username;
      user.ID = userData.Id;
      user.role = userData.Role; // type in Role Enum
      user.forceChangePassword = userData.forceChangePassword;
      user.endpointAuthorizations = userData.EndpointAuthorizations;
      user.portainerAuthorizations = userData.PortainerAuthorizations;

      // Initialize user theme base on UserTheme from database
      const userTheme = userData.ThemeSettings ? userData.ThemeSettings.color : 'auto';
      if (userTheme === 'auto' || !userTheme) {
        ThemeManager.autoTheme();
      } else {
        ThemeManager.setTheme(userTheme);
      }

      LocalStorage.storeUserId(userData.Id);
    }

    // To avoid creating divergence between CE and EE
    // isAdmin checks if the user is a portainer admin or edge admin
    function isAdmin() {
      return isAdminHelperFunc({ Role: user.role });
    }

    // To avoid creating divergence between CE and EE
    // isPureAdmin checks if the user is portainer admin only
    function isPureAdmin() {
      return isPureAdminHelperFunc({ Role: user.role });
    }

    function hasAuthorizations(authorizations) {
      const endpointId = EndpointProvider.endpointID();
      if (isAdmin()) {
        return true;
      }
      if (!user.endpointAuthorizations || !user.endpointAuthorizations[endpointId]) {
        return false;
      }
      const userEndpointAuthorizations = user.endpointAuthorizations[endpointId];
      return authorizations.some((authorization) => userEndpointAuthorizations[authorization]);
    }

    function redirectIfUnauthorized(authorizations) {
      const authorized = hasAuthorizations(authorizations);
      if (!authorized) {
        $state.go('portainer.home');
      }
    }
  },
]);
