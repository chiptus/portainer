import { EnvironmentId } from '@/portainer/environments/types';

import { SidebarItem } from '../SidebarItem';

interface Props {
  environmentId: EnvironmentId;
}

export function NomadSidebar({ environmentId }: Props) {
  return (
    <>
      <SidebarItem
        to="nomad.dashboard"
        params={{ endpointId: environmentId }}
        iconClass="fa-tachometer-alt fa-fw"
        label="Dashboard"
      />
      <SidebarItem
        to="nomad.jobs"
        params={{ endpointId: environmentId }}
        iconClass="fa-th-list fa-fw"
        label="Nomad Jobs"
      />
    </>
  );
}
