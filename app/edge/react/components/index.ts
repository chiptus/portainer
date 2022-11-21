import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { EnvironmentActions } from '@/react/edge/edge-stacks/ItemView/EnvironmentActions';
import { ActionStatus } from '@/react/edge/edge-stacks/ItemView/ActionStatus';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { EdgeAsyncIntervalsForm } from '@/react/edge/components/EdgeAsyncIntervalsForm';
import { EdgeCheckinIntervalField } from '@/react/edge/components/EdgeCheckInIntervalField';
import { EdgeScriptForm } from '@/react/edge/components/EdgeScriptForm';
import { EdgeGroupsSelector } from '@/react/edge/edge-stacks/components/EdgeGroupsSelector';
import { EdgeStackDeploymentTypeSelector } from '@/react/edge/edge-stacks/components/EdgeStackDeploymentTypeSelector';
import { PrivateRegistryFieldset } from '@/react/edge/edge-stacks/components/PrivateRegistryFieldset';

export const componentsModule = angular
  .module('portainer.edge.react.components', [])
  .component(
    'edgeStackEnvironmentActions',
    r2a(withUIRouter(withReactQuery(EnvironmentActions)), [
      'environment',
      'edgeStackId',
    ])
  )
  .component(
    'edgeStackActionStatus',
    r2a(withReactQuery(ActionStatus), ['edgeStackId', 'environmentId'])
  )
  .component(
    'edgeGroupsSelector',
    r2a(EdgeGroupsSelector, ['items', 'onChange', 'value'])
  )
  .component(
    'edgeScriptForm',
    r2a(withReactQuery(EdgeScriptForm), [
      'edgeInfo',
      'commands',
      'isNomadTokenVisible',
      'hideAsyncMode',
    ])
  )
  .component(
    'edgeCheckinIntervalField',
    r2a(withReactQuery(EdgeCheckinIntervalField), [
      'value',
      'onChange',
      'isDefaultHidden',
      'tooltip',
      'label',
      'readonly',
      'size',
    ])
  )
  .component(
    'edgeAsyncIntervalsForm',
    r2a(withReactQuery(EdgeAsyncIntervalsForm), [
      'values',
      'onChange',
      'isDefaultHidden',
      'readonly',
      'fieldSettings',
    ])
  )
  .component(
    'edgeStackDeploymentTypeSelector',
    r2a(EdgeStackDeploymentTypeSelector, [
      'value',
      'onChange',
      'hasDockerEndpoint',
      'hasKubeEndpoint',
      'hasNomadEndpoint',
    ])
  )
  .component(
    'privateRegistryFieldset',
    r2a(PrivateRegistryFieldset, [
      'value',
      'registries',
      'onChange',
      'formInvalid',
      'errorMessage',
      'onSelect',
      'isActive',
      'clearRegistries',
      'method',
    ])
  ).name;
