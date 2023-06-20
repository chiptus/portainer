import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { SettingsFDO } from '@/react/portainer/settings/EdgeComputeView/SettingsFDO';
import { SettingsOpenAMT } from '@/react/portainer/settings/EdgeComputeView/SettingsOpenAMT';
import { InternalAuth } from '@/react/portainer/settings/AuthenticationView/InternalAuth';
import { ExperimentalFeatures } from '@/react/portainer/settings/SettingsView/ExperimentalFeatures';
import { ApplicationSettingsPanel } from '@/react/portainer/settings/SettingsView/ApplicationSettingsPanel';
import { KubeSettingsPanel } from '@/react/portainer/settings/SettingsView/KubeSettingsPanel';

export const settingsModule = angular
  .module('portainer.app.react.components.settings', [])
  .component(
    'settingsFdo',
    r2a(withUIRouter(withReactQuery(SettingsFDO)), ['onSubmit', 'settings'])
  )
  .component('settingsOpenAmt', r2a(SettingsOpenAMT, ['onSubmit', 'settings']))
  .component(
    'settingsExperimentalFeatures',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ExperimentalFeatures))), [])
  )
  .component(
    'internalAuth',
    r2a(InternalAuth, ['onSaveSettings', 'isLoading', 'value', 'onChange'])
  )
  .component(
    'applicationSettingsPanel',
    r2a(withReactQuery(ApplicationSettingsPanel), ['onSuccess'])
  )
  .component(
    'kubeSettingsPanel',
    r2a(withUIRouter(withReactQuery(KubeSettingsPanel)), [])
  ).name;
