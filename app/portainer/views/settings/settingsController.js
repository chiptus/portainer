import angular from 'angular';

import { FeatureId } from '@/react/portainer/feature-flags/enums';
import { options } from '@/react/portainer/settings/SettingsView/backup-options';

angular.module('portainer.app').controller('SettingsController', [
  '$scope',
  'Notifications',
  'SettingsService',
  'StateManager',
  'BackupService',
  'FileSaver',
  function ($scope, Notifications, SettingsService, StateManager, BackupService, FileSaver) {
    $scope.s3BackupFeatureId = FeatureId.S3_BACKUP_SETTING;

    $scope.backupOptions = options;

    $scope.state = {
      isDemo: false,
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

    $scope.BACKUP_FORM_TYPES = { S3: 's3', FILE: 'file' };

    $scope.formValues = {
      customLogo: false,
      customLoginBanner: false,
      GlobalDeploymentOptions: {
        hideAddWithForm: false,
        perEnvOverride: false,
        hideWebEditor: false,
        hideFileUpload: false,
      },
      KubeconfigExpiry: undefined,
      HelmRepositoryURL: undefined,
      BlackListedLabels: [],
      labelName: '',
      labelValue: '',
      enableTelemetry: false,
      passwordProtect: false,
      password: '',
      scheduleAutomaticBackups: true,
      cronRule: '',
      accessKeyId: '',
      secretAccessKey: '',
      region: '',
      bucketName: '',
      s3CompatibleHost: '',
      backupFormType: $scope.BACKUP_FORM_TYPES.FILE,
    };

    $scope.initialFormValues = {};

    $scope.onToggleEnableTelemetry = function onToggleEnableTelemetry(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.enableTelemetry = checked;
      });
    };

    $scope.onToggleCustomLogo = function onToggleCustomLogo(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.customLogo = checked;
      });
    };

    $scope.onToggleCustomLoginBanner = function onToggleCustomLoginBanner(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.customLoginBanner = checked;
      });
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

    $scope.onToggleAutoBackups = function onToggleAutoBackups(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.scheduleAutomaticBackups = checked;
      });
    };

    $scope.onBackupOptionsChange = function (type, limited) {
      $scope.formValues.backupFormType = type;
      $scope.state.featureLimited = limited;
    };

    $scope.onChangeCheckInInterval = function (interval) {
      $scope.$evalAsync(() => {
        var settings = $scope.settings;
        settings.EdgeAgentCheckinInterval = interval;
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

    $scope.downloadBackup = function () {
      const payload = {};
      if ($scope.formValues.passwordProtect) {
        payload.password = $scope.formValues.password;
      }

      $scope.state.backupInProgress = true;

      BackupService.downloadBackup(payload)
        .then(function success(data) {
          const downloadData = new Blob([data.file], { type: 'application/gzip' });
          FileSaver.saveAs(downloadData, data.name);
          Notifications.success('Success', 'Backup successfully downloaded');
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to download backup');
        })
        .finally(function final() {
          $scope.state.backupInProgress = false;
        });
    };

    // only update the values from the app settings widget. In future separate the api endpoints
    $scope.saveApplicationSettings = function () {
      const appSettingsPayload = {
        SnapshotInterval: $scope.settings.SnapshotInterval,
        LogoURL: $scope.formValues.customLogo ? $scope.settings.LogoURL : '',
        EnableTelemetry: $scope.formValues.enableTelemetry,
        CustomLoginBanner: $scope.formValues.customLoginBanner ? $scope.settings.CustomLoginBanner : '',
        TemplatesURL: $scope.settings.TemplatesURL,
        EdgeAgentCheckinInterval: $scope.settings.EdgeAgentCheckinInterval,
      };

      $scope.state.actionInProgress = true;
      updateSettings(appSettingsPayload, 'Application settings updated');
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

    $scope.saveS3BackupSettings = function () {
      const payload = getS3SettingsPayload();
      BackupService.saveS3Settings(payload)
        .then(function success() {
          Notifications.success('Success', 'S3 Backup settings successfully saved');
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to save S3 backup settings');
        })
        .finally(function final() {
          $scope.state.backupInProgress = false;
        });
    };

    $scope.exportBackup = function () {
      const payload = getS3SettingsPayload();
      BackupService.exportBackup(payload)
        .then(function success() {
          Notifications.success('Success', 'Exported backup to S3 successfully');
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to export backup to S3');
        })
        .finally(function final() {
          $scope.state.backupInProgress = false;
        });
    };

    function getS3SettingsPayload() {
      return {
        Password: $scope.formValues.passwordProtectS3 ? $scope.formValues.passwordS3 : '',
        CronRule: $scope.formValues.scheduleAutomaticBackups ? $scope.formValues.cronRule : '',
        AccessKeyID: $scope.formValues.accessKeyId,
        SecretAccessKey: $scope.formValues.secretAccessKey,
        Region: $scope.formValues.region,
        BucketName: $scope.formValues.bucketName,
        S3CompatibleHost: $scope.formValues.s3CompatibleHost,
      };
    }

    function updateSettings(settings, successMessage = 'Settings updated') {
      // ignore CloudApiKeys to avoid overriding them
      //
      // it is not ideal solution as API still accepts CloudAPIKeys
      // which may override the cloud provider API keys
      settings.CloudApiKeys = undefined;
      SettingsService.update(settings)
        .then(function success(response) {
          Notifications.success('Success', successMessage);
          StateManager.updateLogo(settings.LogoURL);
          StateManager.updateSnapshotInterval(settings.SnapshotInterval);
          StateManager.updateEnableTelemetry(settings.EnableTelemetry);
          $scope.initialFormValues.enableTelemetry = response.EnableTelemetry;
          $scope.formValues.BlackListedLabels = response.BlackListedLabels;
          // trigger an event to update the deployment options for the react based sidebar
          const event = new CustomEvent('portainer:deploymentOptionsUpdated');
          document.dispatchEvent(event);
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to update settings');
        })
        .finally(function final() {
          $scope.state.kubeSettingsActionInProgress = false;
          $scope.state.actionInProgress = false;
        });
    }

    function initView() {
      const state = StateManager.getState();
      $scope.state.isDemo = state.application.demoEnvironment.enabled;

      BackupService.getS3Settings()
        .then(function success(data) {
          $scope.formValues.passwordS3 = data.Password;
          $scope.formValues.cronRule = data.CronRule;
          $scope.formValues.accessKeyId = data.AccessKeyID;
          $scope.formValues.secretAccessKey = data.SecretAccessKey;
          $scope.formValues.region = data.Region;
          $scope.formValues.bucketName = data.BucketName;
          $scope.formValues.s3CompatibleHost = data.S3CompatibleHost;

          $scope.formValues.scheduleAutomaticBackups = $scope.formValues.cronRule ? true : false;
          $scope.formValues.passwordProtectS3 = $scope.formValues.passwordS3 ? true : false;
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve S3 backup settings');
        });

      SettingsService.settings()
        .then(function success(data) {
          var settings = data;
          $scope.settings = settings;

          if (settings.LogoURL !== '') {
            $scope.formValues.customLogo = true;
          }

          if (settings.CustomLoginBanner !== '') {
            $scope.formValues.customLoginBanner = true;
          }

          if (settings.GlobalDeploymentOptions) {
            $scope.formValues.GlobalDeploymentOptions = settings.GlobalDeploymentOptions;
          }

          $scope.initialFormValues.enableTelemetry = settings.EnableTelemetry;

          $scope.formValues.enableTelemetry = settings.EnableTelemetry;
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
