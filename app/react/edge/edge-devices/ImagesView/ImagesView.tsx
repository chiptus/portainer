import { useCurrentStateAndParams } from '@uirouter/react';
import { List } from 'lucide-react';

import { useEnvironment } from '@/react/portainer/environments/queries';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import { columns } from '@/react/docker/images/ListView/ImagesDatatable/columns';
import { EdgeDeviceViewsHeader } from '@/react/edge/components/EdgeDeviceViewsHeader';
import { NoSnapshotAvailablePanel } from '@/react/edge/components/NoSnapshotAvailablePanel';
import { ImagesListResponse } from '@/react/docker/images/queries/useImages';

import { Datatable } from '@@/datatables';
import { createPersistedStore } from '@@/datatables/types';
import { useTableState } from '@@/datatables/useTableState';

import { ImagesDatatableActions } from './ImagesDatatableActions';

const storageKey = 'edge_stack_images';
const settingsStore = createPersistedStore(storageKey, 'created');

export function ImagesView() {
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

  const transformedImages: Array<ImagesListResponse> = images.map(
    (image) =>
      ({
        used: containers.some((c) => image.RepoTags.includes(c.Image)),
        created: image.Created,
        id: image.Id,
        size: image.VirtualSize,
        tags: image.RepoTags,
      }) satisfies ImagesListResponse
  );

  return (
    <>
      <EdgeDeviceViewsHeader
        title="Images"
        breadcrumbs={breadcrumbs}
        environment={environment}
      />

      <Datatable<(typeof transformedImages)[0]>
        title="Images"
        titleIcon={List}
        renderTableActions={(selectedRows) => (
          <ImagesDatatableActions
            selectedItems={selectedRows}
            endpointId={environment.Id}
          />
        )}
        dataset={transformedImages}
        settingsManager={tableState}
        columns={columns}
        emptyContentLabel="No images found"
        isRowSelectable={(row) => !row.original.used}
      />
    </>
  );
}
