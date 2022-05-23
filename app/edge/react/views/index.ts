import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { EdgeDevicesView } from '@/react/edge/edge-devices/ListView';

export const viewsModule = angular
  .module('portainer.edge.react.views', [])
  .component('edgeDevicesView', r2a(EdgeDevicesView, [])).name;
