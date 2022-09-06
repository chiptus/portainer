import angular from 'angular';

import { r2a } from '@/react-tools/react2angular';
import { StackContainersDatatable } from '@/react/docker/stacks/ItemView/StackContainersDatatable';
import { StackImageStatus } from '@/react/docker/stacks/ListView/StackImageStatus';
import { ContainerQuickActions } from '@/react/docker/containers/components/ContainerQuickActions';
import { ImageStatus } from '@/react/docker/components/ImageStatus';
import { TemplateListDropdownAngular } from '@/react/docker/app-templates/TemplateListDropdown';
import { TemplateListSortAngular } from '@/react/docker/app-templates/TemplateListSort';

export const componentsModule = angular
  .module('portainer.docker.react.components', [])
  .component(
    'stackImageStatus',
    r2a(StackImageStatus, ['stackId', 'environmentId'])
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
  .component('imageStatus', r2a(ImageStatus, ['imageName', 'environmentId']))
  .component('templateListDropdown', TemplateListDropdownAngular)
  .component('templateListSort', TemplateListSortAngular)
  .component(
    'stackContainersDatatable',
    r2a(StackContainersDatatable, ['environment', 'stackName'])
  ).name;
