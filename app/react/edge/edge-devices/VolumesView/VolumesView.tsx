import { useCurrentStateAndParams } from '@uirouter/react';
import { Database } from 'lucide-react';

import { useEnvironment } from '@/react/portainer/environments/queries';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import { RowProvider } from '@/react/docker/volumes/ListView/VolumesDatatable/RowContext';
import { name } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/name';
import { stackName } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/stackName';
import { driver } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/driver';
import { mountPoint } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/mountpoint';
import { created } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/created';
import { EdgeDeviceViewsHeader } from '@/react/edge/components/EdgeDeviceViewsHeader';
import { NoSnapshotAvailablePanel } from '@/react/edge/components/NoSnapshotAvailablePanel';

import { Datatable } from '@@/datatables/Datatable';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';

import { VolumesDatatableActions } from './VolumesDatatableActions';

const storageKey = 'edge_stack_volumes';
const settingsStore = createPersistedStore(storageKey, 'created');

export const columns = [name, stackName, driver, mountPoint, created];

export function VolumesView() {
  const tableState = useTableState(settingsStore, storageKey);

  const {
    params: { environmentId },
  } = useCurrentStateAndParams();

  if (!environmentId) {
    throw new Error('Missing environmentId parameter');
  }

  const { data: environment } = useEnvironment(environmentId);
  const { data: snapshot } = useDockerSnapshot(environmentId);

  if (!environment) {
    return null;
  }

  const breadcrumbs = [
    { label: 'Edge Devices', link: 'edge.devices' },
    {
      label: environment.Name,
      link: 'edge.browse.dashboard',
      linkParams: { environmentId },
    },
    { label: 'Volumes' },
  ];

  if (!snapshot) {
    return (
      <>
        <EdgeDeviceViewsHeader
          title="Volumes"
          breadcrumbs={breadcrumbs}
          environment={environment}
        />

        <NoSnapshotAvailablePanel />
      </>
    );
  }

  const { Volumes: volumes, Containers: containers } = snapshot;

  const transformedVolumes = volumes.map((v) => ({
    ...v,
    dangling: containers.every((c) => c.Mounts.some((m) => m.Name !== v.Id)),
  }));

  return (
    <>
      <EdgeDeviceViewsHeader
        title="Volumes"
        breadcrumbs={breadcrumbs}
        environment={environment}
      />

      <RowProvider context={{ environment }}>
        <Datatable
          title="Volumes"
          titleIcon={Database}
          renderTableActions={(selectedRows) => (
            <VolumesDatatableActions
              selectedItems={selectedRows}
              endpointId={environment.Id}
            />
          )}
          settingsManager={tableState}
          dataset={transformedVolumes}
          columns={columns}
          emptyContentLabel="No volumes found"
          isRowSelectable={(row) => row.original.dangling}
        />
      </RowProvider>
    </>
  );
}
