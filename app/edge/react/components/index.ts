import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { EnvironmentActions } from '@/react/edge/edge-stacks/ItemView/EnvironmentActions';
import { ActionStatus } from '@/react/edge/edge-stacks/ItemView/ActionStatus';

export const componentsModule = angular
  .module('portainer.edge.react.components', [])
  .component(
    'edgeStackEnvironmentActions',
    r2a(EnvironmentActions, ['environment', 'edgeStackId'])
  )
  .component(
    'edgeStackActionStatus',
    r2a(ActionStatus, ['edgeStackId', 'environmentId'])
  ).name;
