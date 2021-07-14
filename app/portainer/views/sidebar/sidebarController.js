angular.module('portainer.app').controller('SidebarController', [
  '$rootScope',
  '$q',
  '$scope',
  '$transitions',
  'StateManager',
  'Notifications',
  'Authentication',
  'UserService',
  'EndpointProvider',
  function ($rootScope, $q, $scope, $transitions, StateManager, Notifications, Authentication, UserService, EndpointProvider) {
    function checkPermissions(memberships) {
      var isLeader = false;
      angular.forEach(memberships, function (membership) {
        if (membership.Role === 1) {
          isLeader = true;
        }
      });
      $scope.isTeamLeader = isLeader;
    }

    function isClusterAdmin() {
      return Authentication.isAdmin();
    }

    function isEndpointAdmin() {
      return Authentication.hasAuthorizations(['EndpointResourcesAccess']);
    }

    async function initView() {
      $scope.uiVersion = StateManager.getState().application.version;
      //$scope.edition = StateManager.getState().application.edition;
      $scope.edition = 'Business Edition';
      $scope.logo = StateManager.getState().application.logo;

      $scope.endpointId = EndpointProvider.endpointID();
      $scope.showStacks = shouldShowStacks();

      let userDetails = Authentication.getUserDetails();
      const isAdmin = isClusterAdmin();
      $scope.isAdmin = isAdmin;
      $scope.isEndpointAdmin = isEndpointAdmin();

      $q.when(!isAdmin ? UserService.userMemberships(userDetails.ID) : [])
        .then(function success(data) {
          checkPermissions(data);
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve user memberships');
        });
    }

    initView();

    function shouldShowStacks() {
      if (isClusterAdmin() || isEndpointAdmin()) {
        return true;
      }

      const endpoint = EndpointProvider.currentEndpoint();
      if (!endpoint || !endpoint.SecuritySettings) {
        return false;
      }

      return endpoint.SecuritySettings.allowStackManagementForRegularUsers;
    }

    $transitions.onEnter({}, async () => {
      $scope.endpointId = EndpointProvider.endpointID();
      $scope.showStacks = shouldShowStacks();
      $scope.isAdmin = isClusterAdmin();
      $scope.isEndpointAdmin = isEndpointAdmin();

      if ($scope.applicationState.endpoint.name) {
        document.title = `${$rootScope.defaultTitle} | ${$scope.applicationState.endpoint.name}`;
      }
    });
  },
]);
