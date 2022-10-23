import { useCurrentStateAndParams } from '@uirouter/react';

import { useEnvironment } from '@/react/portainer/environments/queries';
import { Datatable } from '@/react/components/datatables/Datatable';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import { createStore } from '@/react/docker/images/ListView/ImagesDatatable/datatable-store';
import { RowProvider } from '@/react/docker/images/ListView/ImagesDatatable/RowContext';
import { id } from '@/react/docker/images/ListView/ImagesDatatable/columns/id';
import { tags } from '@/react/docker/images/ListView/ImagesDatatable/columns/tags';
import { size } from '@/react/docker/images/ListView/ImagesDatatable/columns/size';
import { created } from '@/react/docker/images/ListView/ImagesDatatable/columns/created';
import { EdgeDeviceViewsHeader } from '@/react/edge/components/EdgeDeviceViewsHeader';
import { NoSnapshotAvailablePanel } from '@/react/edge/components/NoSnapshotAvailablePanel';

import { ImagesDatatableActions } from './ImagesDatatableActions';

const storageKey = 'edge_stack_images';
const useStore = createStore(storageKey);

export const columns = [id, tags, size, created];

export function ImagesView() {
  const {
    params: { environmentId },
  } = useCurrentStateAndParams();

  if (!environmentId) {
    throw new Error('Missing environmentId parameter');
  }

  const settings = useStore();
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
    { label: 'Images' },
  ];

  if (!snapshot) {
    return (
      <>
        <EdgeDeviceViewsHeader
          title="Images"
          breadcrumbs={breadcrumbs}
          environment={environment}
        />

        <NoSnapshotAvailablePanel />
      </>
    );
  }

  const { Images: images, Containers: containers } = snapshot;

  const transformedImages = images.map((image) => ({
    ...image,
    Used: containers.some((c) => image.RepoTags.includes(c.Image)),
  }));

  return (
    <>
      <EdgeDeviceViewsHeader
        title="Images"
        breadcrumbs={breadcrumbs}
        environment={environment}
        snapshot={snapshot}
      />

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
          dataset={transformedImages}
          columns={columns}
          settingsStore={settings}
          emptyContentLabel="No images found"
          isRowSelectable={(row) => !row.original.Used}
        />
      </RowProvider>
    </>
  );
}
