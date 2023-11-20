import { getCurrentUser } from '../users/queries/useLoadCurrentUser';
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

    var service = {};
    var user = {};

    service.init = init;
    service.OAuthLogin = OAuthLogin;
    service.login = login;
    service.logout = logout;
    service.isAuthenticated = isAuthenticated;
    service.getUserDetails = getUserDetails;
    service.isAdmin = isAdmin;
    service.hasAuthorizations = hasAuthorizations;
    service.redirectIfUnauthorized = redirectIfUnauthorized;

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

    async function loadUserData() {
      const userData = await getCurrentUser();
      user.username = userData.Username;
      user.ID = userData.Id;
      user.role = userData.Role;
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

    function isAdmin() {
      if (user.role === 1) {
        return true;
      }
      return false;
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

    if (process.env.NODE_ENV === 'development') {
      window.login = loginAsync;
    }

    return service;
  },
]);
