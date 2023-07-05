import angular from 'angular';

import { HomeView } from '@/react/portainer/HomeView';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { r2a } from '@/react-tools/react2angular';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { CreateAccessToken } from '@/react/portainer/account/CreateAccessTokenView';
import { EdgeComputeSettingsView } from '@/react/portainer/settings/EdgeComputeView/EdgeComputeSettingsView';
import { CloudView } from '@/react/portainer/settings/sharedCredentials/CloudView';
import { CreateCredentialView } from '@/react/portainer/settings/sharedCredentials/CreateCredentialsView';
import { EditCredentialView } from '@/react/portainer/settings/sharedCredentials/EditCredentialView';
import { withI18nSuspense } from '@/react-tools/withI18nSuspense';
import { NotificationsView } from '@/react/portainer/notifications/NotificationsView';
import { EdgeAutoCreateScriptView } from '@/react/portainer/environments/EdgeAutoCreateScriptView';
import { ListView as EnvironmentsListView } from '@/react/portainer/environments/ListView';
import { BackupSettingsPanel } from '@/react/portainer/settings/SettingsView/BackupSettingsView/BackupSettingsPanel';
import { CreateView as licenseCreateView } from '@/react/portainer/licenses/CreateView/CreateView';

import { wizardModule } from './wizard';
import { teamsModule } from './teams';
import { updateSchedulesModule } from './update-schedules';
import { accountViews } from './account';

export const viewsModule = angular
  .module('portainer.app.react.views', [
    wizardModule,
    teamsModule,
    updateSchedulesModule,
    accountViews,
  ])
  .component(
    'homeView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(HomeView))), [])
  )
  .component(
    'edgeAutoCreateScriptView',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(EdgeAutoCreateScriptView))),
      []
    )
  )
  .component(
    'createAccessToken',
    r2a(withI18nSuspense(withUIRouter(CreateAccessToken)), [
      'onSubmit',
      'onError',
    ])
  )
  .component(
    'settingsEdgeCompute',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(EdgeComputeSettingsView))),
      ['onSubmit', 'settings']
    )
  )
  .component(
    'settingsSharedCredentialsView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(CloudView))), [])
  )
  .component(
    'addSharedCredentialsView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(CreateCredentialView))), [])
  )
  .component(
    'editSharedCredentialsView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EditCredentialView))), [])
  )
  .component(
    'backupSettingsPanel',
    r2a(withUIRouter(withReactQuery(withCurrentUser(BackupSettingsPanel))), [])
  )
  .component(
    'notifications',
    r2a(withUIRouter(withReactQuery(withCurrentUser(NotificationsView))), [])
  )
  .component(
    'environmentsListView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EnvironmentsListView))), [])
  )
  .component(
    'licenseCreateView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(licenseCreateView))), [])
  ).name;
