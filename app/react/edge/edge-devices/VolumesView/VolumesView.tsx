import { useCurrentStateAndParams } from '@uirouter/react';

import { useEnvironment } from '@/portainer/environments/queries';
import { Datatable } from '@/react/components/datatables/Datatable';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import { createStore } from '@/react/docker/volumes/ListView/VolumesDatatable/datatable-store';
import { RowProvider } from '@/react/docker/volumes/ListView/VolumesDatatable/RowContext';
import { id } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/id';
import { stackName } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/stackName';
import { driver } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/driver';
import { mountpoint } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/mountpoint';
import { created } from '@/react/docker/volumes/ListView/VolumesDatatable/columns/created';
import { SnapshotBrowsingPanel } from '@/react/edge/components/SnapshotBrowsingPanel';

import { PageHeader } from '@@/PageHeader';

import { NoSnapshotAvailablePanel } from '../NoSnapshotAvailablePanel';

import { VolumesDatatableActions } from './VolumesDatatableActions';

const storageKey = 'edge_stack_volumes';
const useStore = createStore(storageKey);

export const columns = [id, stackName, driver, mountpoint, created];

export function VolumesView() {
  const {
    params: { environmentId },
  } = useCurrentStateAndParams();

  const settings = useStore();
  const environmentQuery = useEnvironment(environmentId);
  const snapshotQuery = useDockerSnapshot(environmentId);

  if (!environmentId) {
    throw new Error('Missing environmentId parameter');
  }

  if (!environmentQuery.data) {
    return null;
  }

  const { data: environment } = environmentQuery;

  if (!snapshotQuery.data) {
    return (
      <>
        <Header name={environment.Name} environmentId={environmentId} />

        <NoSnapshotAvailablePanel />
      </>
    );
  }

  const {
    data: {
      Volumes: volumes,
      SnapshotTime: snapshotTime,
      Containers: containers,
    },
  } = snapshotQuery;

  volumes.forEach((v) => {
    const volume = v;
    const used = containers.some((c) => c.Mounts.some((m) => m.Name === v.Id));
    if (used) {
      volume.Used = true;
    }
  });

  return (
    <>
      <Header name={environment.Name} environmentId={environmentId} />

      <div className="row">
        <div className="col-sm-12">
          <SnapshotBrowsingPanel snapshotTime={snapshotTime} />
        </div>
      </div>

      <RowProvider context={{ environment }}>
        <Datatable
          titleOptions={{
            icon: 'fa-cubes',
            title: 'Volumes',
          }}
          renderTableActions={(selectedRows) => (
            <VolumesDatatableActions
              selectedItems={selectedRows}
              endpointId={environment.Id}
            />
          )}
          storageKey={storageKey}
          dataset={volumes}
          columns={columns}
          settingsStore={settings}
          emptyContentLabel="No volumes found"
          isRowSelectable={(row) => !row.original.Used}
        />
      </RowProvider>
    </>
  );
}

function Header({
  name,
  environmentId,
}: {
  name: string;
  environmentId: string;
}) {
  return (
    <PageHeader
      title="Volumes"
      breadcrumbs={[
        { label: 'Edge Devices', link: 'edge.devices' },
        {
          label: name,
          link: 'edge.browse.dashboard',
          linkParams: { environmentId },
        },
        { label: 'Volumes' },
      ]}
      reload
    />
  );
}
