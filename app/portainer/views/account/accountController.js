angular.module('portainer.app').controller('AccountController', [
  '$scope',
  '$state',
  'Authentication',
  'UserService',
  'Notifications',
  'SettingsService',
  'ModalService',
  'StateManager',
  function ($scope, $state, Authentication, UserService, Notifications, SettingsService, ModalService, StateManager) {
    $scope.formValues = {
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    };

    $scope.updatePassword = async function () {
      const confirmed = await ModalService.confirmChangePassword();
      if (confirmed) {
        try {
          await UserService.updateUserPassword($scope.userID, $scope.formValues.currentPassword, $scope.formValues.newPassword);
          Notifications.success('Success', 'Password successfully updated');
          StateManager.resetPasswordChangeSkips($scope.userID);
          $scope.forceChangePassword = false;
          $state.go('portainer.logout');
        } catch (err) {
          Notifications.error('Failure', err, err.msg);
        }
      }
    };

    $scope.skipPasswordChange = async function () {
      try {
        if ($scope.userCanSkip()) {
          StateManager.setPasswordChangeSkipped($scope.userID);
          $scope.forceChangePassword = false;
          $state.go('portainer.home');
        }
      } catch (err) {
        Notifications.error('Failure', err, err.msg);
      }
    };

    $scope.userCanSkip = function () {
      return $scope.timesPasswordChangeSkipped < 2;
    };

    this.uiCanExit = (newTransition) => {
      if ($scope.userRole === 1 && newTransition.to().name === 'portainer.settings.authentication') {
        return true;
      }
      if (newTransition.to().name === 'portainer.logout') {
        return true;
      }
      if ($scope.forceChangePassword) {
        ModalService.confirmForceChangePassword();
      }
      return !$scope.forceChangePassword;
    };

    $scope.uiCanExit = () => {
      return this.uiCanExit();
    };

    $scope.removeAction = (selectedTokens) => {
      const msg = 'Do you want to remove the selected access token(s)? Any script or application using these tokens will no longer be able to invoke the Portainer API.';

      ModalService.confirmDeletion(msg, (confirmed) => {
        if (!confirmed) {
          return;
        }
        let actionCount = selectedTokens.length;
        selectedTokens.forEach((token) => {
          UserService.deleteAccessToken($scope.userID, token.id)
            .then(() => {
              Notifications.success('Token successfully removed');
              var index = $scope.tokens.indexOf(token);
              $scope.tokens.splice(index, 1);
            })
            .catch((err) => {
              Notifications.error('Failure', err, 'Unable to remove token');
            })
            .finally(() => {
              --actionCount;
              if (actionCount === 0) {
                $state.reload();
              }
            });
        });
      });
    };

    function initView() {
      const state = StateManager.getState();
      const userDetails = Authentication.getUserDetails();
      $scope.userID = userDetails.ID;
      $scope.userRole = Authentication.getUserDetails().role;
      $scope.forceChangePassword = userDetails.forceChangePassword;

      if (state.application.demoEnvironment.enabled) {
        $scope.isDemoUser = state.application.demoEnvironment.users.includes($scope.userID);
      }

      SettingsService.publicSettings()
        .then(function success(data) {
          $scope.AuthenticationMethod = data.AuthenticationMethod;

          if (state.UI.requiredPasswordLength && state.UI.requiredPasswordLength !== data.RequiredPasswordLength) {
            StateManager.clearPasswordChangeSkips($scope.userID);
          }

          $scope.timesPasswordChangeSkipped =
            state.UI.timesPasswordChangeSkipped && state.UI.timesPasswordChangeSkipped[$scope.userID] ? state.UI.timesPasswordChangeSkipped[$scope.userID] : 0;

          $scope.requiredPasswordLength = data.RequiredPasswordLength;
          StateManager.setRequiredPasswordLength(data.RequiredPasswordLength);
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve application settings');
        });
      UserService.getAccessTokens($scope.userID)
        .then(function success(data) {
          $scope.tokens = data;
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve user tokens');
        });
    }

    initView();
  },
]);
