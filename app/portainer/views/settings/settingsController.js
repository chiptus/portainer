import angular from 'angular';

angular.module('portainer.app').controller('SettingsController', [
  '$scope',
  'Notifications',
  'SettingsService',
  'StateManager',
  function ($scope, Notifications, SettingsService, StateManager) {
    $scope.updateSettings = updateSettings;
    $scope.handleSuccess = handleSuccess;

    $scope.state = {
      actionInProgress: false,
      availableKubeconfigExpiryOptions: [
        {
          key: '1 day',
          value: '24h',
        },
        {
          key: '7 days',
          value: `${24 * 7}h`,
        },
        {
          key: '30 days',
          value: `${24 * 30}h`,
        },
        {
          key: '1 year',
          value: `${24 * 30 * 12}h`,
        },
        {
          key: 'No expiry',
          value: '0',
        },
      ],
      backupInProgress: false,
      featureLimited: false,
    };

    $scope.formValues = {
      GlobalDeploymentOptions: {
        hideAddWithForm: false,
        perEnvOverride: false,
        hideWebEditor: false,
        hideFileUpload: false,
        requireNoteOnApplications: false,
        minApplicationNoteLength: '',
      },
      KubeconfigExpiry: undefined,
      HelmRepositoryURL: undefined,
      BlackListedLabels: [],
      labelName: '',
      labelValue: '',
    };

    $scope.onToggleAddWithForm = function onToggleAddWithForm(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.GlobalDeploymentOptions.hideAddWithForm = checked;
        $scope.formValues.GlobalDeploymentOptions.hideWebEditor = false;
        $scope.formValues.GlobalDeploymentOptions.hideFileUpload = false;
        if (checked) {
          $scope.formValues.GlobalDeploymentOptions.hideWebEditor = true;
          $scope.formValues.GlobalDeploymentOptions.hideFileUpload = true;
        }
      });
    };

    $scope.onToggleNoteOnApplications = function onToggleNoteOnApplications(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.GlobalDeploymentOptions.requireNoteOnApplications = checked;
      });
    };

    $scope.onTogglePerEnvOverride = function onTogglePerEnvOverride(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.GlobalDeploymentOptions.perEnvOverride = checked;
      });
    };

    $scope.onToggleHideWebEditor = function onToggleHideWebEditor(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.GlobalDeploymentOptions.hideWebEditor = !checked;
      });
    };

    $scope.onToggleHideFileUpload = function onToggleHideFileUpload(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.GlobalDeploymentOptions.hideFileUpload = !checked;
      });
    };

    $scope.removeFilteredContainerLabel = function (index) {
      const filteredSettings = $scope.formValues.BlackListedLabels.filter((_, i) => i !== index);
      const filteredSettingsPayload = { BlackListedLabels: filteredSettings };
      updateSettings(filteredSettingsPayload, 'Hidden container settings updated');
    };

    $scope.addFilteredContainerLabel = function () {
      var label = {
        name: $scope.formValues.labelName,
        value: $scope.formValues.labelValue,
      };

      const filteredSettings = [...$scope.formValues.BlackListedLabels, label];
      const filteredSettingsPayload = { BlackListedLabels: filteredSettings };
      updateSettings(filteredSettingsPayload, 'Hidden container settings updated');
    };

    // only update the values from the kube settings widget. In future separate the api endpoints
    $scope.saveKubernetesSettings = function () {
      const kubeSettingsPayload = {
        KubeconfigExpiry: $scope.formValues.KubeconfigExpiry,
        HelmRepositoryURL: $scope.formValues.HelmRepositoryURL,
        GlobalDeploymentOptions: $scope.formValues.GlobalDeploymentOptions,
      };

      $scope.state.kubeSettingsActionInProgress = true;
      updateSettings(kubeSettingsPayload, 'Kubernetes settings updated');
    };

    function updateSettings(settings, successMessage = 'Settings updated') {
      // ignore CloudApiKeys to avoid overriding them
      //
      // it is not ideal solution as API still accepts CloudAPIKeys
      // which may override the cloud provider API keys
      settings.CloudApiKeys = undefined;
      return SettingsService.update(settings)
        .then(function success(settings) {
          Notifications.success('Success', successMessage);
          handleSuccess(settings);
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to update settings');
        })
        .finally(function final() {
          $scope.state.kubeSettingsActionInProgress = false;
          $scope.state.actionInProgress = false;
        });
    }

    function handleSuccess(settings) {
      if (settings) {
        StateManager.updateLogo(settings.LogoURL);
        StateManager.updateSnapshotInterval(settings.SnapshotInterval);
        StateManager.updateEnableTelemetry(settings.EnableTelemetry);
        $scope.formValues.BlackListedLabels = settings.BlackListedLabels;
      }

      // trigger an event to update the deployment options for the react based sidebar
      const event = new CustomEvent('portainer:deploymentOptionsUpdated');
      document.dispatchEvent(event);
    }

    function initView() {
      SettingsService.settings()
        .then(function success(data) {
          var settings = data;
          $scope.settings = settings;

          if (settings.GlobalDeploymentOptions) {
            $scope.formValues.GlobalDeploymentOptions = settings.GlobalDeploymentOptions;
          }

          $scope.formValues.KubeconfigExpiry = settings.KubeconfigExpiry;
          $scope.formValues.HelmRepositoryURL = settings.HelmRepositoryURL;
          $scope.formValues.BlackListedLabels = settings.BlackListedLabels;
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve application settings');
        });
    }

    initView();
  },
]);
