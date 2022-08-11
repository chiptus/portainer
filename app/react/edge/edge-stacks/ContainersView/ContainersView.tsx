import { useCurrentStateAndParams } from '@uirouter/react';

import { useEnvironment } from '@/portainer/environments/queries';
import { name } from '@/react/docker/containers/ListView/ContainersDatatable/columns/name';
import { state } from '@/react/docker/containers/ListView/ContainersDatatable/columns/state';
import { image } from '@/react/docker/containers/ListView/ContainersDatatable/columns/image';
import { created } from '@/react/docker/containers/ListView/ContainersDatatable/columns/created';
import { ip } from '@/react/docker/containers/ListView/ContainersDatatable/columns/ip';
import { host } from '@/react/docker/containers/ListView/ContainersDatatable/columns/host';
import { ports } from '@/react/docker/containers/ListView/ContainersDatatable/columns/ports';
import { Datatable } from '@/react/components/datatables/Datatable';
import { useDockerSnapshotContainers } from '@/react/docker/queries/useDockerSnapshotContainers';
import { createStore } from '@/react/docker/containers/ListView/ContainersDatatable/datatable-store';

import { PageHeader } from '@@/PageHeader';

import { useEdgeStack } from '../queries/useEdgeStack';

const storageKey = 'edge_stack_containers';
const useStore = createStore(storageKey);
const columns = [name, state, image, created, ip, host, ports];

export function ContainersView() {
  const {
    params: { environmentId, stackId },
  } = useCurrentStateAndParams();

  const settings = useStore();

  const environmentQuery = useEnvironment(environmentId);

  const edgeStackQuery = useEdgeStack(stackId);

  const containersQuery = useDockerSnapshotContainers(environmentId, {
    edgeStackId: stackId,
  });

  if (!environmentId || !stackId) {
    throw new Error('Missing environmentId or stackId parameters');
  }

  if (!environmentQuery.data || !edgeStackQuery.data || !containersQuery.data) {
    return null;
  }

  const { data: environment } = environmentQuery;
  const { data: edgeStack } = edgeStackQuery;

  return (
    <>
      <PageHeader
        title="Containers"
        breadcrumbs={[
          { label: 'Edge Stacks', link: 'edge.stacks' },
          { label: edgeStack.Name, link: 'edge.stacks.edit', linkParams: {} },
          { label: environment.Name },
          { label: 'Containers' },
        ]}
      />

      <Datatable
        titleOptions={{
          icon: 'fa-cubes',
          title: 'Containers',
        }}
        storageKey={storageKey}
        dataset={containersQuery.data}
        columns={columns}
        settingsStore={settings}
        disableSelect
        emptyContentLabel="No containers found"
      />
    </>
  );
}
