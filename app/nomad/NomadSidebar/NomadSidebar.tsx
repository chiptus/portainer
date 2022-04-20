import { r2a } from '@/react-tools/react2angular';
import { SidebarMenuItem } from '@/portainer/components/sidebar/SidebarMenuItem';

interface Props {
  environmentId: string;
}

export function NomadSidebar({ environmentId }: Props) {
  return (
    <>
      <SidebarMenuItem
        path="nomad.dashboard"
        pathParams={{ endpointId: environmentId }}
        iconClass="fa-tachometer-alt fa-fw"
        className="sidebar-list"
        itemName="Dashboard"
        data-cy="nomadSidebar-dashboard"
      >
        Dashboard
      </SidebarMenuItem>
      <SidebarMenuItem
        path="nomad.jobs"
        pathParams={{ endpointId: environmentId }}
        iconClass="fa-th-list fa-fw"
        className="sidebar-list"
        itemName="Jobs"
        data-cy="nomadSidebar-jobs"
      >
        Nomad Jobs
      </SidebarMenuItem>
    </>
  );
}

export const NomadSidebarAngular = r2a(NomadSidebar, ['environmentId']);
