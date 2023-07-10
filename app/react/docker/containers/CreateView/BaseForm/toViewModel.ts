import { parseAccessControlFormData } from '@/react/portainer/access-control/utils';
import { ResourceControlOwnership } from '@/react/portainer/access-control/types';
import { UserId } from '@/portainer/users/types';

import { ContainerResponse } from '../../queries/container';

import { toViewModel as toPortsMappingViewModel } from './PortsMappingField.viewModel';
import { Values } from './BaseForm';

export function toViewModel(
  isAdmin: boolean,
  currentUserId: UserId,
  config: ContainerResponse
): Omit<Values, 'enableWebhook' | 'image' | 'nodeName'> {
  // accessControl shouldn't be copied to new container

  const accessControl = parseAccessControlFormData(isAdmin, currentUserId);

  if (config.Portainer?.ResourceControl?.Public) {
    accessControl.ownership = ResourceControlOwnership.PUBLIC;
  }

  return {
    accessControl,
    name: config.Name ? config.Name.replace('/', '') : '',
    alwaysPull: true,
    autoRemove: config.HostConfig?.AutoRemove || false,
    ports: toPortsMappingViewModel(config.HostConfig?.PortBindings || {}),
    publishAllPorts: config.HostConfig?.PublishAllPorts || false,
  };
}

export function getDefaultViewModel(
  isAdmin: boolean,
  currentUserId: UserId
): Omit<Values, 'enableWebhook' | 'image' | 'nodeName'> {
  const accessControl = parseAccessControlFormData(isAdmin, currentUserId);

  return {
    accessControl,
    name: '',
    alwaysPull: true,
    autoRemove: false,
    ports: [],
    publishAllPorts: false,
  };
}
