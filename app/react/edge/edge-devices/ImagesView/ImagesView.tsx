import { useCurrentStateAndParams } from '@uirouter/react';
import { List } from 'react-feather';
import { useStore } from 'zustand';

import { useEnvironment } from '@/react/portainer/environments/queries';
import { Datatable } from '@/react/components/datatables/Datatable';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import { RowProvider } from '@/react/docker/images/ListView/ImagesDatatable/RowContext';
import { id } from '@/react/docker/images/ListView/ImagesDatatable/columns/id';
import { tags } from '@/react/docker/images/ListView/ImagesDatatable/columns/tags';
import { size } from '@/react/docker/images/ListView/ImagesDatatable/columns/size';
import { created } from '@/react/docker/images/ListView/ImagesDatatable/columns/created';
import { EdgeDeviceViewsHeader } from '@/react/edge/components/EdgeDeviceViewsHeader';
import { NoSnapshotAvailablePanel } from '@/react/edge/components/NoSnapshotAvailablePanel';

import { useSearchBarState } from '@@/datatables/SearchBar';
import { createPersistedStore } from '@@/datatables/types';

import { ImagesDatatableActions } from './ImagesDatatableActions';

export const columns = [id, tags, size, created];

const storageKey = 'edge_stack_images';
const settingsStore = createPersistedStore(storageKey);

export function ImagesView() {
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
          title="Images"
          titleIcon={List}
          renderTableActions={(selectedRows) => (
            <ImagesDatatableActions
              selectedItems={selectedRows}
              endpointId={environment.Id}
            />
          )}
          dataset={transformedImages}
          columns={columns}
          emptyContentLabel="No images found"
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
