import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { withControlledInput } from '@/react-tools/withControlledInput';
import { StackContainersDatatable } from '@/react/docker/stacks/ItemView/StackContainersDatatable';
import { StackImageStatus } from '@/react/docker/stacks/ListView/StackImageStatus';
import { ContainerQuickActions } from '@/react/docker/containers/components/ContainerQuickActions';
import { ImageStatus } from '@/react/docker/components/ImageStatus';
import { TemplateListDropdown } from '@/react/docker/app-templates/TemplateListDropdown';
import { TemplateListSort } from '@/react/docker/app-templates/TemplateListSort';
import { Gpu } from '@/react/docker/containers/CreateView/Gpu';
import { withCurrentUser } from '@/react-tools/withCurrentUser';
import { withReactQuery } from '@/react-tools/withReactQuery';
import { withUIRouter } from '@/react-tools/withUIRouter';
import { DockerfileDetails } from '@/react/docker/images/ItemView/DockerfileDetails';
import { HealthStatus } from '@/react/docker/containers/ItemView/HealthStatus';
import { GpusList } from '@/react/docker/host/SetupView/GpusList';
import { GpusInsights } from '@/react/docker/host/SetupView/GpusInsights';
import { InsightsBox } from '@/react/components/InsightsBox';

export const componentsModule = angular
  .module('portainer.docker.react.components', [])
  .component('dockerfileDetails', r2a(DockerfileDetails, ['image']))
  .component('dockerHealthStatus', r2a(HealthStatus, ['health']))
  .component(
    'stackImageStatus',
    r2a(withReactQuery(StackImageStatus), ['stackId', 'environmentId'])
  )
  .component(
    'containerQuickActions',
    r2a(withUIRouter(withCurrentUser(ContainerQuickActions)), [
      'containerId',
      'nodeName',
      'state',
      'status',
      'taskId',
    ])
  )
  .component(
    'imageStatus',
    r2a(withReactQuery(ImageStatus), [
      'environmentId',
      'resourceId',
      'resourceType',
      'nodeName',
    ])
  )
  .component(
    'templateListDropdown',
    r2a(TemplateListDropdown, ['options', 'onChange', 'placeholder', 'value'])
  )
  .component(
    'templateListSort',
    r2a(TemplateListSort, [
      'options',
      'onChange',
      'onDescending',
      'placeholder',
      'sortByDescending',
      'sortByButton',
      'value',
    ])
  )
  .component(
    'stackContainersDatatable',
    r2a(
      withUIRouter(withReactQuery(withCurrentUser(StackContainersDatatable))),
      ['environment', 'stackName']
    )
  )
  .component(
    'gpu',
    r2a(Gpu, [
      'values',
      'onChange',
      'gpus',
      'usedGpus',
      'usedAllGpus',
      'enableGpuManagement',
    ])
  )
  .component(
    'gpusList',
    r2a(withControlledInput(GpusList), ['value', 'onChange'])
  )
  .component(
    'insightsBox',
    r2a(InsightsBox, [
      'header',
      'content',
      'insightCloseId',
      'type',
      'className',
    ])
  )
  .component('gpusInsights', r2a(GpusInsights, [])).name;
