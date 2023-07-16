import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { EdgeAsyncIntervalsForm } from '@/react/edge/components/EdgeAsyncIntervalsForm';
import { EdgeCheckinIntervalField } from '@/react/edge/components/EdgeCheckInIntervalField';
import { EdgeScriptForm } from '@/react/edge/components/EdgeScriptForm';
import { EdgeGroupsSelector } from '@/react/edge/edge-stacks/components/EdgeGroupsSelector';
import { EdgeStackDeploymentTypeSelector } from '@/react/edge/edge-stacks/components/EdgeStackDeploymentTypeSelector';
import { PrivateRegistryFieldset } from '@/react/edge/edge-stacks/components/PrivateRegistryFieldset';
import { EditEdgeStackForm } from '@/react/edge/edge-stacks/ItemView/EditEdgeStackForm/EditEdgeStackForm';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { EdgeGroupAssociationTable } from '@/react/edge/components/EdgeGroupAssociationTable';
import { AssociatedEdgeEnvironmentsSelector } from '@/react/edge/components/AssociatedEdgeEnvironmentsSelector';
import { EnvironmentsDatatable } from '@/react/edge/edge-stacks/ItemView/EnvironmentsDatatable';
import { EdgeStackStatus } from '@/react/edge/edge-stacks/ListView/EdgeStackStatus';

export const componentsModule = angular
  .module('portainer.edge.react.components', [])
  .component('edgeStacksDatatableStatus', r2a(EdgeStackStatus, ['edgeStack']))
  .component(
    'edgeStackEnvironmentsDatatable',
    r2a(withUIRouter(withReactQuery(EnvironmentsDatatable)), [])
  )
  .component(
    'edgeGroupsSelector',
    r2a(withUIRouter(withReactQuery(EdgeGroupsSelector)), [
      'onChange',
      'value',
      'error',
      'horizontal',
      'isGroupVisible',
    ])
  )
  .component(
    'edgeScriptForm',
    r2a(withReactQuery(EdgeScriptForm), [
      'edgeInfo',
      'commands',
      'isNomadTokenVisible',
      'asyncMode',
      'showMetaFields',
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
      'allowKubeToSelectCompose',
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
  )
  .component(
    'editEdgeStackForm',
    r2a(withUIRouter(withReactQuery(withCurrentUser(EditEdgeStackForm))), [])
  )
  .component(
    'edgeGroupAssociationTable',
    r2a(withReactQuery(EdgeGroupAssociationTable), [
      'emptyContentLabel',
      'onClickRow',
      'query',
      'title',
      'data-cy',
      'hideEnvironmentIds',
    ])
  )
  .component(
    'associatedEdgeEnvironmentsSelector',
    r2a(withReactQuery(AssociatedEdgeEnvironmentsSelector), [
      'onChange',
      'value',
    ])
  ).name;
