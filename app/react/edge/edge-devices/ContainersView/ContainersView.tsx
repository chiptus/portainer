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
import { RowProvider } from '@/react/docker/containers/ListView/ContainersDatatable/RowContext';
import { useEdgeStack } from '@/react/edge/edge-stacks/queries/useEdgeStack';

import { PageHeader } from '@@/PageHeader';
import { TextTip } from '@@/Tip/TextTip';
import { Widget } from '@@/Widget';

const storageKey = 'edge_stack_containers';
const useStore = createStore(storageKey);
const columns = [name, state, image, created, ip, host, ports];

export function ContainersView() {
  const {
    params: { environmentId, edgeStackId },
  } = useCurrentStateAndParams();

  const settings = useStore();

  const environmentQuery = useEnvironment(environmentId);

  const edgeStackQuery = useEdgeStack(edgeStackId);

  const containersQuery = useDockerSnapshotContainers(environmentId, {
    edgeStackId,
  });

  if (!environmentId) {
    throw new Error('Missing environmentId parameter');
  }

  if (
    !environmentQuery.data ||
    !containersQuery.data ||
    (edgeStackId && !edgeStackQuery.data)
  ) {
    return null;
  }

  const { data: environment } = environmentQuery;

  return (
    <>
      <PageHeader
        title="Containers"
        breadcrumbs={[
          { label: 'Edge Devices', link: 'edge.devices' },
          { label: environment.Name },
          { label: 'Containers' },
        ]}
      />

      {edgeStackQuery.data && (
        <div className="row">
          <div className="col-sm-12">
            <Widget>
              <Widget.Body>
                <TextTip color="blue">
                  Containers are filtered by edge stack
                  <span className="ml-px">{edgeStackQuery.data?.Name}</span>.
                </TextTip>
              </Widget.Body>
            </Widget>
          </div>
        </div>
      )}

      <RowProvider context={{ environment }}>
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
      </RowProvider>
    </>
  );
}
