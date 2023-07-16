import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { ContainersView } from '@/react/edge/edge-devices/ContainersView';
import { StacksView } from '@/react/edge/edge-devices/StacksView';
import { ContainerView } from '@/react/edge/edge-devices/ContainerView';
import { DashboardView } from '@/react/edge/edge-devices/DashboardView';
import { ImagesView } from '@/react/edge/edge-devices/ImagesView';
import { VolumesView } from '@/react/edge/edge-devices/VolumesView';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { WaitingRoomView } from '@/react/edge/edge-devices/WaitingRoomView';
import { ListView } from '@/react/edge/edge-stacks/ListView';
import { ItemView as EdgeStackItemView } from '@/react/edge/edge-stacks/ItemView/ItemView';

export const viewsModule = angular
  .module('portainer.edge.react.views', [])
  .component(
    'waitingRoomView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(WaitingRoomView))), [])
  )
  .component(
    'edgeDeviceDashboardView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(DashboardView))), [])
  )
  .component(
    'edgeDeviceStacksView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(StacksView))), [])
  )
  .component(
    'edgeDeviceContainersView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ContainersView))), [])
  )
  .component(
    'edgeDeviceContainerView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ContainerView))), [])
  )
  .component(
    'edgeDeviceImagesView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(ImagesView))), [])
  )
  .component(
    'edgeDeviceVolumesView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(VolumesView))), [])
  )
  .component(
    'edgeStackView',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EdgeStackItemView))), [])
  )
  .component(
    'edgeStacksView',
    r2a(withUIRouter(withCurrentUser(ListView)), [])
  ).name;
