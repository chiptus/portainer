import { useCurrentStateAndParams } from '@uirouter/react';
import { find, forEach } from 'lodash';

import { useEnvironment } from '@/portainer/environments/queries';
import { Datatable } from '@/react/components/datatables/Datatable';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import { createStore } from '@/react/docker/images/ListView/ImagesDatatable/datatable-store';
import { RowProvider } from '@/react/docker/images/ListView/ImagesDatatable/RowContext';
import { id } from '@/react/docker/images/ListView/ImagesDatatable/columns/id';
import { tags } from '@/react/docker/images/ListView/ImagesDatatable/columns/tags';
import { size } from '@/react/docker/images/ListView/ImagesDatatable/columns/size';
import { created } from '@/react/docker/images/ListView/ImagesDatatable/columns/created';
import { SnapshotBrowsingPanel } from '@/react/edge/components/SnapshotBrowsingPanel';

import { PageHeader } from '@@/PageHeader';

import { NoSnapshotAvailablePanel } from '../NoSnapshotAvailablePanel';

import { ImagesDatatableActions } from './ImagesDatatableActions';

const storageKey = 'edge_stack_images';
const useStore = createStore(storageKey);

export const columns = [id, tags, size, created];

export function ImagesView() {
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
      Images: images,
      SnapshotTime: snapshotTime,
      Containers: containers,
    },
  } = snapshotQuery;

  forEach(containers, (container) => {
    const image = find(images, (image) =>
      image.RepoTags.includes(container.Image)
    );
    if (image) {
      image.Used = true;
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
            title: 'Images',
          }}
          renderTableActions={(selectedRows) => (
            <ImagesDatatableActions
              selectedItems={selectedRows}
              endpointId={environment.Id}
            />
          )}
          storageKey={storageKey}
          dataset={images}
          columns={columns}
          settingsStore={settings}
          emptyContentLabel="No images found"
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
      title="Images"
      breadcrumbs={[
        { label: 'Edge Devices', link: 'edge.devices' },
        {
          label: name,
          link: 'edge.browse.dashboard',
          linkParams: { environmentId },
        },
        { label: 'Images' },
      ]}
      reload
    />
  );
}
