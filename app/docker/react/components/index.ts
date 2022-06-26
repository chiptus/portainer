import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { StackImageStatus } from '@/react/docker/stacks/ListView/StackImageStatus';
import { ContainersDatatableContainer } from '@/react/docker/containers/ListView/ContainersDatatable/ContainersDatatableContainer';
import { ContainerQuickActions } from '@/react/docker/containers/components/ContainerQuickActions';
import { ImageStatus } from '@/react/docker/components/ImageStatus';

export const componentsModule = angular
  .module('portainer.docker.react.components', [])
  .component('stackImageStatus', r2a(StackImageStatus, ['stackId']))
  .component(
    'containersDatatable',
    r2a(ContainersDatatableContainer, [
      'endpoint',
      'isAddActionVisible',
      'dataset',
      'onRefresh',
      'isHostColumnVisible',
      'tableKey',
    ])
  )
  .component(
    'containerQuickActions',
    r2a(ContainerQuickActions, [
      'containerId',
      'nodeName',
      'state',
      'status',
      'taskId',
    ])
  )
  .component(
    'imageStatus',
    r2a(ImageStatus, ['imageName', 'environmentId'])
  ).name;
