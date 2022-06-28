import { Clock } from 'react-feather';

import { EnvironmentId } from '@/portainer/environments/types';

import { DashboardLink } from '../items/DashboardLink';
import { SidebarItem } from '../SidebarItem';

interface Props {
  environmentId: EnvironmentId;
}

export function NomadSidebar({ environmentId }: Props) {
  return (
    <>
      <DashboardLink environmentId={environmentId} platformPath="nomad" />
      <SidebarItem
        to="nomad.jobs"
        params={{ endpointId: environmentId }}
        icon={Clock}
        label="Nomad Jobs"
      />
    </>
  );
}
