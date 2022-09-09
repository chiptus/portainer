import angular from 'angular';

import { FeatureId } from '@/portainer/feature-flags/enums';
import { options } from './options';

angular.module('portainer.app').controller('SettingsController', [
  '$scope',
  '$state',
  'Notifications',
  'SettingsService',
  'StateManager',
  'BackupService',
  'FileSaver',
  function ($scope, $state, Notifications, SettingsService, StateManager, BackupService, FileSaver) {
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
      backupFormType: $scope.BACKUP_FORM_TYPES.FILE,
    };

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

    $scope.onToggleAutoBackups = function onToggleAutoBackups(checked) {
      $scope.$evalAsync(() => {
        $scope.formValues.scheduleAutomaticBackups = checked;
      });
    };

    $scope.onBackupOptionsChange = function (type, limited) {
      $scope.formValues.backupFormType = type;
      $scope.state.featureLimited = limited;
    };

    $scope.removeFilteredContainerLabel = function (index) {
      var settings = $scope.settings;
      settings.BlackListedLabels.splice(index, 1);

      updateSettings(settings);
    };

    $scope.addFilteredContainerLabel = function () {
      var settings = $scope.settings;
      var label = {
        name: $scope.formValues.labelName,
        value: $scope.formValues.labelValue,
      };
      settings.BlackListedLabels.push(label);

      updateSettings(settings);
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

    $scope.saveApplicationSettings = function () {
      var settings = $scope.settings;

      if (!$scope.formValues.customLogo) {
        settings.LogoURL = '';
      }

      if (!$scope.formValues.customLoginBanner) {
        settings.CustomLoginBanner = '';
      }

      settings.EnableTelemetry = $scope.formValues.enableTelemetry;

      $scope.state.actionInProgress = true;
      updateSettings(settings);
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
      };
    }

    function updateSettings(settings) {
      // ignore CloudApiKeys to avoid overriding them
      //
      // it is not ideal solution as API still accepts CloudAPIKeys
      // which may override the cloud provider API keys
      settings.CloudApiKeys = undefined;
      SettingsService.update(settings)
        .then(function success() {
          Notifications.success('Success', 'Settings updated');
          StateManager.updateLogo(settings.LogoURL);
          StateManager.updateSnapshotInterval(settings.SnapshotInterval);
          StateManager.updateEnableTelemetry(settings.EnableTelemetry);
          $state.reload();
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to update settings');
        })
        .finally(function final() {
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
          $scope.formValues.enableTelemetry = settings.EnableTelemetry;
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve application settings');
        });
    }

    initView();
  },
]);
