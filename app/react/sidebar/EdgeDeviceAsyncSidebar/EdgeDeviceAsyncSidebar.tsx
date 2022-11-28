import { Box, Database, Layout, List } from 'lucide-react';

import { EnvironmentId } from '@/react/portainer/environments/types';

import { SidebarItem } from '../SidebarItem';

interface Props {
  environmentId: EnvironmentId;
}

export function EdgeDeviceAsyncSidebar({ environmentId }: Props) {
  return (
    <>
      <SidebarItem
        to="edge.browse.dashboard"
        params={{ environmentId }}
        icon={Layout}
        label="Dashboard"
        data-cy="edgeDeviceSidebar-dashboard"
      />

      <SidebarItem
        to="edge.browse.containers"
        params={{ environmentId }}
        icon={Box}
        label="Containers"
        data-cy="edgeDeviceSidebar-containers"
      />

      <SidebarItem
        to="edge.browse.images"
        params={{ environmentId }}
        icon={List}
        label="Images"
        data-cy="edgeDeviceSidebar-images"
      />

      <SidebarItem
        to="edge.browse.volumes"
        params={{ environmentId }}
        icon={Database}
        label="Volumes"
        data-cy="edgeDeviceSidebar-volumes"
      />
    </>
  );
}
