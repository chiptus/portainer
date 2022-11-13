import angular from 'angular';

import { HomeView } from '@/react/portainer/HomeView';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { r2a } from '@/react-tools/react2angular';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { CreateAccessToken } from '@/react/portainer/account/CreateAccessTokenView';
import { EdgeComputeSettingsView } from '@/react/portainer/settings/EdgeComputeView/EdgeComputeSettingsView';
import { CloudView } from '@/react/portainer/settings/cloud/CloudView';
import { CreateCredentialView } from '@/react/portainer/settings/cloud/CreateCredentialsView';
import { EditCredentialView } from '@/react/portainer/settings/cloud/EditCredentialView';
import { withI18nSuspense } from '@/react-tools/withI18nSuspense';

import { wizardModule } from './wizard';
import { teamsModule } from './teams';
import { updateSchedulesModule } from './update-schedules';

export const viewsModule = angular
  .module('portainer.app.react.views', [
    wizardModule,
    teamsModule,
    updateSchedulesModule,
  ])
  .component(
    'homeView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(HomeView))), [])
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
    r2a(withReactQuery(withCurrentUser(EdgeComputeSettingsView)), [
      'onSubmit',
      'settings',
    ])
  )
  .component(
    'settingsCloudView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(CloudView))), [])
  )
  .component(
    'addCloudCredentialView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(CreateCredentialView))), [])
  )
  .component(
    'editCloudCredentialView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EditCredentialView))), [])
  ).name;
