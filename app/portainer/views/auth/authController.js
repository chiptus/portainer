import angular from 'angular';
import uuidv4 from 'uuid/v4';

class AuthenticationController {
  /* @ngInject */
  constructor(
    $async,
    $analytics,
    $scope,
    $state,
    $stateParams,
    $window,
    Authentication,
    UserService,
    EndpointService,
    StateManager,
    Notifications,
    SettingsService,
    URLHelper,
    LocalStorage,
    StatusService,
    LicenseService
  ) {
    this.$async = $async;
    this.$analytics = $analytics;
    this.$scope = $scope;
    this.$state = $state;
    this.$stateParams = $stateParams;
    this.$window = $window;
    this.Authentication = Authentication;
    this.UserService = UserService;
    this.EndpointService = EndpointService;
    this.StateManager = StateManager;
    this.Notifications = Notifications;
    this.SettingsService = SettingsService;
    this.URLHelper = URLHelper;
    this.LocalStorage = LocalStorage;
    this.StatusService = StatusService;
    this.LicenseService = LicenseService;

    this.logo = this.StateManager.getState().application.logo;
    this.formValues = {
      Username: '',
      Password: '',
    };
    this.state = {
      showOAuthLogin: false,
      showStandardLogin: false,
      AuthenticationError: '',
      loginInProgress: true,
      OAuthProvider: '',
    };

    this.checkForEndpointsAsync = this.checkForEndpointsAsync.bind(this);
    this.checkForLicensesAsync = this.checkForLicensesAsync.bind(this);
    this.checkForLatestVersionAsync = this.checkForLatestVersionAsync.bind(this);
    this.postLoginSteps = this.postLoginSteps.bind(this);

    this.oAuthLoginAsync = this.oAuthLoginAsync.bind(this);
    this.internalLoginAsync = this.internalLoginAsync.bind(this);

    this.authenticateUserAsync = this.authenticateUserAsync.bind(this);

    this.manageOauthCodeReturn = this.manageOauthCodeReturn.bind(this);
    this.authEnabledFlowAsync = this.authEnabledFlowAsync.bind(this);
    this.onInit = this.onInit.bind(this);
  }

  /**
   * UTILS FUNCTIONS SECTION
   */

  logout(error) {
    this.Authentication.logout();
    this.state.loginInProgress = false;
    this.generateOAuthLoginURI();
    this.LocalStorage.storeLogoutReason(error);
    this.$window.location.reload();
  }

  error(err, message) {
    this.state.AuthenticationError = message;
    if (!err) {
      err = {};
    }
    this.Notifications.error('Failure', err, message);
    this.state.loginInProgress = false;
  }

  determineOauthProvider(LoginURI) {
    if (LoginURI.indexOf('login.microsoftonline.com') !== -1) {
      return 'Microsoft';
    } else if (LoginURI.indexOf('accounts.google.com') !== -1) {
      return 'Google';
    } else if (LoginURI.indexOf('github.com') !== -1) {
      return 'Github';
    }
    return 'OAuth';
  }

  generateState() {
    const uuid = uuidv4();
    this.LocalStorage.storeLoginStateUUID(uuid);
    return '&state=' + uuid;
  }

  generateOAuthLoginURI() {
    this.OAuthLoginURI = this.state.OAuthLoginURI + this.generateState();
  }

  hasValidState(state) {
    const savedUUID = this.LocalStorage.getLoginStateUUID();
    return savedUUID && state && savedUUID === state;
  }

  /**
   * END UTILS FUNCTIONS SECTION
   */

  /**
   * POST LOGIN STEPS SECTION
   */

  async checkForEndpointsAsync() {
    try {
      const endpoints = await this.EndpointService.endpoints(0, 1);

      if (endpoints.value.length === 0) {
        return 'portainer.wizard';
      }
    } catch (err) {
      this.error(err, 'Unable to retrieve environments');
    }
    return 'portainer.home';
  }

  async checkForLicensesAsync() {
    try {
      const info = await this.LicenseService.info();

      if (!info.valid) {
        return 'portainer.init.license';
      }
    } catch (err) {
      this.error(err, 'Unable to retrieve licenses info');
    }
  }

  async checkForLatestVersionAsync() {
    let versionInfo = {
      UpdateAvailable: false,
      LatestVersion: '',
    };

    try {
      const versionStatus = await this.StatusService.version();
      if (versionStatus.UpdateAvailable) {
        versionInfo.UpdateAvailable = true;
        versionInfo.LatestVersion = versionStatus.LatestVersion;
      }
    } finally {
      this.StateManager.setVersionInfo(versionInfo);
    }
  }

  async postLoginSteps() {
    await this.StateManager.initialize();

    const isAdmin = this.Authentication.isAdmin();
    this.$analytics.setUserRole(isAdmin ? 'admin' : 'standard-user');

    let path = 'portainer.home';
    if (isAdmin) {
      this.LicenseService.resetState();
      path = await this.checkForLicensesAsync();
      if (!path) {
        path = await this.checkForEndpointsAsync();
      }
    }

    if (this.Authentication.getUserDetails().forceChangePassword) {
      path = 'portainer.account';
    }

    this.$state.go(path);
  }
  /**
   * END POST LOGIN STEPS SECTION
   */

  /**
   * LOGIN METHODS SECTION
   */

  async oAuthLoginAsync(code) {
    try {
      await this.Authentication.OAuthLogin(code);
      this.URLHelper.cleanParameters();
    } catch (err) {
      this.error(err, 'Unable to login via OAuth');
    }
  }

  async internalLoginAsync(username, password) {
    await this.Authentication.login(username, password);
    await this.postLoginSteps();
  }

  /**
   * END LOGIN METHODS SECTION
   */

  /**
   * AUTHENTICATE USER SECTION
   */

  async authenticateUserAsync() {
    try {
      var username = this.formValues.Username;
      var password = this.formValues.Password;
      this.state.loginInProgress = true;
      await this.internalLoginAsync(username, password);
    } catch (err) {
      this.error(err, 'Unable to login');
    }
  }

  authenticateUser() {
    return this.$async(this.authenticateUserAsync);
  }

  /**
   * END AUTHENTICATE USER SECTION
   */

  /**
   * ON INIT SECTION
   */
  async manageOauthCodeReturn(code, state) {
    if (this.hasValidState(state)) {
      await this.oAuthLoginAsync(code);
    } else {
      this.error(null, 'Invalid OAuth state, try again.');
    }
  }

  async authEnabledFlowAsync() {
    try {
      const exists = await this.UserService.administratorExists();
      if (!exists) {
        this.$state.go('portainer.init.admin');
      }
    } catch (err) {
      this.error(err, 'Unable to verify administrator account existence');
    }
  }

  toggleStandardLogin() {
    this.state.showStandardLogin = !this.state.showStandardLogin;
  }

  async onInit() {
    try {
      const settings = await this.SettingsService.publicSettings();
      this.state.hideInternalAuth = settings.OAuthHideInternalAuth;
      this.state.showOAuthLogin = settings.AuthenticationMethod === 3;
      this.state.showStandardLogin = !this.state.showOAuthLogin;
      this.state.OAuthLoginURI = settings.OAuthLoginURI;
      this.state.OAuthProvider = this.determineOauthProvider(settings.OAuthLoginURI);

      const code = this.URLHelper.getParameter('code');
      const state = this.URLHelper.getParameter('state');
      if (code && state) {
        await this.manageOauthCodeReturn(code, state);
        this.generateOAuthLoginURI();
        return;
      }
      if (!this.logo) {
        await this.StateManager.initialize();
        this.logo = this.StateManager.getState().application.logo;
      }
      this.generateOAuthLoginURI();

      if (this.$stateParams.logout || this.$stateParams.error) {
        this.logout(this.$stateParams.error);
        return;
      }
      const error = this.LocalStorage.getLogoutReason();
      if (error) {
        this.state.AuthenticationError = error;
        this.LocalStorage.cleanLogoutReason();
      }

      if (this.Authentication.isAuthenticated()) {
        await this.postLoginSteps();
      }
      this.state.loginInProgress = false;

      if (this.state.hideInternalAuth && !this.Authentication.isAuthenticated()) {
        this.$window.location.href = this.OAuthLoginURI;
      }

      await this.authEnabledFlowAsync();
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve public settings');
    }
  }

  $onInit() {
    return this.$async(this.onInit);
  }

  /**
   * END ON INIT SECTION
   */
}

export default AuthenticationController;
angular.module('portainer.app').controller('AuthenticationController', AuthenticationController);
