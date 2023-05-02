import { useCurrentStateAndParams } from '@uirouter/react';
import { Box } from 'lucide-react';

import { useEnvironment } from '@/react/portainer/environments/queries';
import { name } from '@/react/docker/containers/ListView/ContainersDatatable/columns/name';
import { state } from '@/react/docker/containers/ListView/ContainersDatatable/columns/state';
import { created } from '@/react/docker/containers/ListView/ContainersDatatable/columns/created';
import { ip } from '@/react/docker/containers/ListView/ContainersDatatable/columns/ip';
import { host } from '@/react/docker/containers/ListView/ContainersDatatable/columns/host';
import { ports } from '@/react/docker/containers/ListView/ContainersDatatable/columns/ports';
import { useDockerSnapshotContainers } from '@/react/docker/queries/useDockerSnapshotContainers';
import { createStore } from '@/react/docker/containers/ListView/ContainersDatatable/datatable-store';
import { RowProvider } from '@/react/docker/containers/ListView/ContainersDatatable/RowContext';
import { useEdgeStack } from '@/react/edge/edge-stacks/queries/useEdgeStack';
import { NoSnapshotAvailablePanel } from '@/react/edge/components/NoSnapshotAvailablePanel';
import { EdgeDeviceViewsHeader } from '@/react/edge/components/EdgeDeviceViewsHeader';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';

import { Datatable } from '@@/datatables/Datatable';
import { TextTip } from '@@/Tip/TextTip';
import { Widget } from '@@/Widget';
import { TableSettingsProvider } from '@@/datatables/useTableSettings';
import { useTableState } from '@@/datatables/useTableState';

import { image } from './image-column';
import { ContainersDatatableActions } from './ContainersDatatableActions';

const storageKey = 'edge_stack_containers';
const settingsStore = createStore(storageKey);
const columns = [name, state, image, created, ip, host, ports];

export function ContainersView() {
  const {
    params: { environmentId, edgeStackId },
  } = useCurrentStateAndParams();

  const environmentQuery = useEnvironment(environmentId);

  const tableState = useTableState(settingsStore, storageKey);

  const edgeStackQuery = useEdgeStack(edgeStackId);

  const containersQuery = useDockerSnapshotContainers(environmentId, {
    edgeStackId,
  });

  const snapshotQuery = useDockerSnapshot(environmentId);

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
  const { data: snapshot } = snapshotQuery;

  const breadcrumbs = [
    { label: 'Edge Devices', link: 'edge.devices' },
    {
      label: environment.Name,
      link: 'edge.browse.dashboard',
      linkParams: { environmentId },
    },
    { label: 'Containers' },
  ];

  if (!snapshot) {
    return (
      <>
        <EdgeDeviceViewsHeader
          title="Containers"
          breadcrumbs={breadcrumbs}
          environment={environment}
        />

        <NoSnapshotAvailablePanel />
      </>
    );
  }

  return (
    <>
      <EdgeDeviceViewsHeader
        title="Containers"
        breadcrumbs={breadcrumbs}
        environment={environment}
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
        <TableSettingsProvider settings={settingsStore}>
          <Datatable
            settingsManager={tableState}
            titleIcon={Box}
            title="Containers"
            renderTableActions={(selectedRows) => (
              <ContainersDatatableActions
                selectedItems={selectedRows}
                endpointId={environment.Id}
              />
            )}
            dataset={containersQuery.data}
            isRowSelectable={(row) =>
              !row.original.Labels['io.portainer.agent']
            }
            columns={columns}
            emptyContentLabel="No containers found"
          />
        </TableSettingsProvider>
      </RowProvider>
    </>
  );
}
