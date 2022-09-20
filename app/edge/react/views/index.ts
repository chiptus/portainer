import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { EdgeDevicesView } from '@/react/edge/edge-devices/ListView';
import { ContainersView } from '@/react/edge/edge-devices/ContainersView';
import { ContainerView } from '@/react/edge/edge-devices/ContainerView';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';

export const viewsModule = angular
  .module('portainer.edge.react.views', [])
  .component(
    'edgeDevicesView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EdgeDevicesView))), [])
  )
  .component(
    'edgeStackEnvironmentContainersView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ContainersView))), [])
  )
  .component(
    'edgeStackEnvironmentContainerView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ContainerView))), [])
  ).name;
