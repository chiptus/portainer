import { useCurrentStateAndParams } from '@uirouter/react';
import { Layers } from 'lucide-react';

import { useEnvironment } from '@/react/portainer/environments/queries';
import { createStore } from '@/react/docker/containers/ListView/ContainersDatatable/datatable-store';
import { RowProvider } from '@/react/docker/containers/ListView/ContainersDatatable/RowContext';
import { useEdgeStacks } from '@/react/edge/edge-stacks/queries/useEdgeStacks';
import { NoSnapshotAvailablePanel } from '@/react/edge/components/NoSnapshotAvailablePanel';
import { EdgeDeviceViewsHeader } from '@/react/edge/components/EdgeDeviceViewsHeader';
import { useDockerSnapshot } from '@/react/docker/queries/useDockerSnapshot';
import { useStacks } from '@/react/docker/stacks/queries/useStacks';
import { filterUniqueContainersBasedOnStack } from '@/react/docker/snapshots/utils';

import { Datatable } from '@@/datatables';
import { TableSettingsProvider } from '@@/datatables/useTableSettings';
import { useTableState } from '@@/datatables/useTableState';

import { StacksDatatableActions } from './StacksDatatableActions';
import { StackInAsyncSnapshot } from './types';
import { useColumns } from './columns';

const storageKey = 'edge_stack_stacks';
const settingsStore = createStore(storageKey);

export function StacksView() {
  const {
    params: { environmentId, edgeStackId },
  } = useCurrentStateAndParams();

  const columns = useColumns();
  const environmentQuery = useEnvironment(environmentId);

  const tableState = useTableState(settingsStore, storageKey);

  const edgeStackQuery = useEdgeStacks();

  const stackQuery = useStacks();

  const snapshotQuery = useDockerSnapshot(environmentId);

  if (!environmentId) {
    throw new Error('Missing environmentId parameter');
  }

  if (
    !environmentQuery.data ||
    // !containersQuery.data ||
    (edgeStackId && !edgeStackQuery.data)
  ) {
    return null;
  }

  const { data: environment } = environmentQuery;
  const { data: snapshot } = snapshotQuery;
  const { data: edgeStacks = [] } = edgeStackQuery;
  const { data: stacks = [] } = stackQuery;

  const breadcrumbs = [
    { label: 'Edge Devices', link: 'edge.devices' },
    {
      label: environment.Name,
      link: 'edge.browse.dashboard',
      linkParams: { environmentId },
    },
    { label: 'Stacks' },
  ];

  if (!snapshot) {
    return (
      <>
        <EdgeDeviceViewsHeader
          title="Stacks"
          breadcrumbs={breadcrumbs}
          environment={environment}
        />

        <NoSnapshotAvailablePanel />
      </>
    );
  }

  const unqiueContainers = filterUniqueContainersBasedOnStack(
    snapshot.Containers
  );

  // filter stacks by labels
  const stacksInAsyncSnapshot = unqiueContainers.map((item) => {
    const stackInAsyncSnapshot: StackInAsyncSnapshot = {
      ...item,
      Metadata: {},
    };

    // determine if container mataches edge stack
    const edgeStack = edgeStacks.some(
      (el) => item.StackName === `edge_${el.Name}`
    );
    stackInAsyncSnapshot.Metadata.isEdgeStack = edgeStack; // control column

    // determine if container mataches stack
    const stack = stacks.find((el) => item.StackName === el.Name);
    stackInAsyncSnapshot.Metadata.isStack = !!stack; // control column
    stackInAsyncSnapshot.Metadata.stackId = stack?.Id;

    // if container deployed by stack does not match either edge
    // stack or stack, we consider it as external stack
    if (
      !stackInAsyncSnapshot.Metadata.isEdgeStack &&
      !stackInAsyncSnapshot.Metadata.isStack
    ) {
      stackInAsyncSnapshot.Metadata.isExternalStack = true;
    }

    return stackInAsyncSnapshot;
  });

  return (
    <>
      <EdgeDeviceViewsHeader
        title="Stacks"
        breadcrumbs={breadcrumbs}
        environment={environment}
      />

      <RowProvider context={{ environment }}>
        <TableSettingsProvider settings={settingsStore}>
          <Datatable
            settingsManager={tableState}
            titleIcon={Layers}
            title="Stacks"
            renderTableActions={(selectedRows) => (
              <StacksDatatableActions
                selectedItems={selectedRows}
                endpointId={environment.Id}
              />
            )}
            dataset={stacksInAsyncSnapshot}
            isRowSelectable={(row) =>
              !(
                row.original.Labels['io.portainer.agent'] ||
                row.original.Metadata.isEdgeStack ||
                row.original.Metadata.isExternalStack
              )
            }
            columns={columns}
            emptyContentLabel="No stacks found"
          />
        </TableSettingsProvider>
      </RowProvider>
    </>
  );
}
