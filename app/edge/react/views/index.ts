import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { EdgeDevicesView } from '@/react/edge/edge-devices/ListView';
import { ContainersView } from '@/react/edge/edge-devices/ContainersView';
import { ContainerView } from '@/react/edge/edge-devices/ContainerView';

export const viewsModule = angular
  .module('portainer.edge.react.views', [])
  .component('edgeDevicesView', r2a(EdgeDevicesView, []))
  .component('edgeDeviceContainersView', r2a(ContainersView, []))
  .component('edgeDeviceContainerView', r2a(ContainerView, [])).name;
