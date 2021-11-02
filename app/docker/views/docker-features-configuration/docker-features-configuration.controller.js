import { HIDE_AUTO_UPDATE_WINDOW } from 'Portainer/feature-flags/feature-ids';

export default class DockerFeaturesConfigurationController {
  /* @ngInject */
  constructor($analytics, $async, EndpointService, Notifications, StateManager) {
    this.$analytics = $analytics;
    this.$async = $async;
    this.EndpointService = EndpointService;
    this.Notifications = Notifications;
    this.StateManager = StateManager;

    this.limitedFeature = HIDE_AUTO_UPDATE_WINDOW;

    this.formValues = {
      enableHostManagementFeatures: false,
      allowVolumeBrowserForRegularUsers: false,
      disableBindMountsForRegularUsers: false,
      disablePrivilegedModeForRegularUsers: false,
      disableHostNamespaceForRegularUsers: false,
      disableStackManagementForRegularUsers: false,
      disableDeviceMappingForRegularUsers: false,
      disableContainerCapabilitiesForRegularUsers: false,
      disableSysctlSettingForRegularUsers: false,
    };

    this.isAgent = false;

    this.state = {
      actionInProgress: false,
    };

    this.save = this.save.bind(this);
  }

  isContainerEditDisabled() {
    const {
      disableBindMountsForRegularUsers,
      disableHostNamespaceForRegularUsers,
      disablePrivilegedModeForRegularUsers,
      disableDeviceMappingForRegularUsers,
      disableContainerCapabilitiesForRegularUsers,
      disableSysctlSettingForRegularUsers,
    } = this.formValues;
    return (
      disableBindMountsForRegularUsers ||
      disableHostNamespaceForRegularUsers ||
      disablePrivilegedModeForRegularUsers ||
      disableDeviceMappingForRegularUsers ||
      disableContainerCapabilitiesForRegularUsers ||
      disableSysctlSettingForRegularUsers
    );
  }

  async save() {
    return this.$async(async () => {
      try {
        this.state.actionInProgress = true;
        const securitySettings = {
          enableHostManagementFeatures: this.formValues.enableHostManagementFeatures,
          allowBindMountsForRegularUsers: !this.formValues.disableBindMountsForRegularUsers,
          allowPrivilegedModeForRegularUsers: !this.formValues.disablePrivilegedModeForRegularUsers,
          allowVolumeBrowserForRegularUsers: this.formValues.allowVolumeBrowserForRegularUsers,
          allowHostNamespaceForRegularUsers: !this.formValues.disableHostNamespaceForRegularUsers,
          allowDeviceMappingForRegularUsers: !this.formValues.disableDeviceMappingForRegularUsers,
          allowStackManagementForRegularUsers: !this.formValues.disableStackManagementForRegularUsers,
          allowContainerCapabilitiesForRegularUsers: !this.formValues.disableContainerCapabilitiesForRegularUsers,
          allowSysctlSettingForRegularUsers: !this.formValues.disableSysctlSettingForRegularUsers,
        };
        const settings = {
          securitySettings,
          changeWindow: this.state.autoUpdateSettings,
        };

        await this.EndpointService.updateSettings(this.endpoint.Id, settings);

        this.endpoint.SecuritySettings = settings.securitySettings;
        this.endpoint.ChangeWindow = settings.changeWindow;

        // Timezone is only for Analytics, not for API payload
        this.$analytics.eventTrack('time-window-create', {
          category: 'docker',
          metadata: {
            'Start-time': settings.changeWindow.StartTime,
            'End-time': settings.changeWindow.EndTime,
            'Time-zone': this.state.timeZone,
          },
        });

        this.Notifications.success('Saved settings successfully');
      } catch (e) {
        this.Notifications.error('Failure', e, 'Failed saving settings');
      }
      this.state.actionInProgress = false;
    });
  }

  checkAgent() {
    const applicationState = this.StateManager.getState();
    return applicationState.endpoint.mode.agentProxy;
  }

  $onInit() {
    const securitySettings = this.endpoint.SecuritySettings;

    const isAgent = this.checkAgent();
    this.isAgent = isAgent;

    this.formValues = {
      enableHostManagementFeatures: isAgent && securitySettings.enableHostManagementFeatures,
      allowVolumeBrowserForRegularUsers: isAgent && securitySettings.allowVolumeBrowserForRegularUsers,
      disableBindMountsForRegularUsers: !securitySettings.allowBindMountsForRegularUsers,
      disablePrivilegedModeForRegularUsers: !securitySettings.allowPrivilegedModeForRegularUsers,
      disableHostNamespaceForRegularUsers: !securitySettings.allowHostNamespaceForRegularUsers,
      disableDeviceMappingForRegularUsers: !securitySettings.allowDeviceMappingForRegularUsers,
      disableStackManagementForRegularUsers: !securitySettings.allowStackManagementForRegularUsers,
      disableContainerCapabilitiesForRegularUsers: !securitySettings.allowContainerCapabilitiesForRegularUsers,
      disableSysctlSettingForRegularUsers: !securitySettings.allowSysctlSettingForRegularUsers,
    };
    this.state = {
      autoUpdateSettings: this.endpoint.ChangeWindow,
      timeZone: '',
    };
  }
}
