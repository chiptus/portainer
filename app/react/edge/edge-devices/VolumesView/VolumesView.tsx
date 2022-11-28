import { useCurrentStateAndParams } from '@uirouter/react';
import { useStore } from 'zustand';
import { Database } from 'lucide-react';

import { useEnvironment } from '@/react/portainer/environments/queries';
import { Datatable } from '@/react/components/datatables/Datatable';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import { RowProvider } from '@/react/docker/volumes/ListView/VolumesDatatable/RowContext';
import { id } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/id';
import { stackName } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/stackName';
import { driver } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/driver';
import { mountpoint } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/mountpoint';
import { created } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/created';
import { EdgeDeviceViewsHeader } from '@/react/edge/components/EdgeDeviceViewsHeader';
import { NoSnapshotAvailablePanel } from '@/react/edge/components/NoSnapshotAvailablePanel';

import { useSearchBarState } from '@@/datatables/SearchBar';
import { createPersistedStore } from '@@/datatables/types';

import { VolumesDatatableActions } from './VolumesDatatableActions';

const storageKey = 'edge_stack_volumes';
const settingsStore = createPersistedStore(storageKey);

export const columns = [id, stackName, driver, mountpoint, created];

export function VolumesView() {
  const settings = useStore(settingsStore);
  const [search, setSearch] = useSearchBarState(storageKey);

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
    Used: containers.some((c) => c.Mounts.some((m) => m.Name === v.Id)),
  }));

  return (
    <>
      <EdgeDeviceViewsHeader
        title="Volumes"
        breadcrumbs={breadcrumbs}
        environment={environment}
        snapshot={snapshot}
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
          dataset={transformedVolumes}
          columns={columns}
          emptyContentLabel="No volumes found"
          isRowSelectable={(row) => !row.original.Used}
          initialPageSize={settings.pageSize}
          onPageSizeChange={settings.setPageSize}
          initialSortBy={settings.sortBy}
          onSortByChange={settings.setSortBy}
          searchValue={search}
          onSearchChange={setSearch}
        />
      </RowProvider>
    </>
  );
}
