import angular from 'angular';
import { buildDefaultValue as buildTunnelDefaultValue } from '@/react/portainer/common/PortainerTunnelAddrField';
import { buildDefaultValue as buildApiUrlDefaultValue } from '@/react/portainer/common/PortainerUrlField';
import { configureFDO } from '@/portainer/hostmanagement/fdo/fdo.service';
import { configureAMT } from 'Portainer/hostmanagement/open-amt/open-amt.service';

angular.module('portainer.app').controller('SettingsEdgeComputeController', SettingsEdgeComputeController);

/* @ngInject */
export default function SettingsEdgeComputeController($q, $async, $state, Notifications, SettingsService, StateManager) {
  var ctrl = this;

  this.onSubmitEdgeCompute = async function (settings) {
    try {
      let mtlsValues = {
        UseSeparateCert: false,
        CaCert: '',
        Cert: '',
        Key: '',
      };

      if (settings.Edge.MTLS.UseSeparateCert) {
        const caCert = settings.Edge.MTLS.CaCertFile ? settings.Edge.MTLS.CaCertFile.text() : '';
        const cert = settings.Edge.MTLS.CertFile ? settings.Edge.MTLS.CertFile.text() : '';
        const key = settings.Edge.MTLS.KeyFile ? settings.Edge.MTLS.KeyFile.text() : '';

        mtlsValues = {
          ...settings.Edge.MTLS,
          CaCert: await caCert,
          Cert: await cert,
          Key: await key,
        };
      }

      settings.Edge = {
        ...settings.Edge,
        MTLS: mtlsValues,
      };

      await SettingsService.update(settings);
      Notifications.success('Success', 'Settings updated');
      StateManager.updateEnableEdgeComputeFeatures(settings.EnableEdgeComputeFeatures);
      $state.reload();
    } catch (err) {
      Notifications.error('Failure', err, 'Unable to update settings');
    }
  };

  this.onSubmitOpenAMT = async function (formValues) {
    try {
      await configureAMT(formValues);
      Notifications.success('Success', `OpenAMT successfully ${formValues.enabled ? 'enabled' : 'disabled'}`);
      $state.reload();
    } catch (err) {
      Notifications.error('Failure', err, 'Failed applying changes');
    }
  };

  this.onSubmitFDO = async function (formValues) {
    try {
      await configureFDO(formValues);
      Notifications.success('Success', `FDO successfully ${formValues.enabled ? 'enabled' : 'disabled'}`);
      $state.reload();
    } catch (err) {
      Notifications.error('Failure', err, 'Failed applying changes');
    }
  };

  function initView() {
    $async(async () => {
      try {
        const defaultApiServerURL = buildApiUrlDefaultValue();
        const defaultTunnelServerAddress = buildTunnelDefaultValue();

        const settings = await SettingsService.settings();
        ctrl.settings = {
          ...settings,
          EdgePortainerUrl: settings.EdgePortainerUrl || defaultApiServerURL,
          Edge: {
            ...settings.Edge,
            TunnelServerAddress: settings.Edge.TunnelServerAddress || defaultTunnelServerAddress,
          },
        };
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to retrieve application settings');
      }
    });
  }

  initView();
}
