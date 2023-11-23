import { Node, Endpoints } from 'kubernetes-types/core/v1';
import { HardDrive, Plus, Trash2 } from 'lucide-react';
import { useMemo } from 'react';
import { useRouter } from '@uirouter/react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { createStore } from '@/react/kubernetes/datatables/default-kube-datatable-store';
import {
  Authorized,
  useAuthorizations,
  useCurrentUser,
} from '@/react/hooks/useUser';
import { IndexOptional } from '@/react/kubernetes/configs/types';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { pluralize } from '@/portainer/helpers/strings';
import {
  EnvironmentId,
  K8sDistributionType,
} from '@/react/portainer/environments/types';
import { notifySuccess } from '@/portainer/services/notifications';
import { useCloudCredential } from '@/react/portainer/settings/sharedCredentials/cloudSettings.service';
import { useAnalytics } from '@/react/hooks/useAnalytics';

import { Datatable, TableSettingsMenu } from '@@/datatables';
import { useTableState } from '@@/datatables/useTableState';
import { Button } from '@@/buttons';
import { confirmDelete } from '@@/modals/confirm';
import { Link } from '@@/Link';
import { TextTip } from '@@/Tip/TextTip';
import { TableSettingsMenuAutoRefresh } from '@@/datatables/TableSettingsMenuAutoRefresh';

import { useNodesQuery, useRemoveNodesMutation } from '../nodes.service';
import { useKubernetesEndpointsQuery } from '../../kubernetesEndpoint.service';

import { getColumns } from './columns';
import { NodeRowData } from './types';
import { getInternalNodeIpAddress, getRole } from './utils';

const storageKey = 'k8sNodesDatatable';
const settingsStore = createStore(storageKey);

export function NodesDatatable() {
  const tableState = useTableState(settingsStore, storageKey);
  const environmentId = useEnvironmentId();
  const { data: nodes, ...nodesQuery } = useNodesQuery(environmentId, {
    autoRefreshRate: tableState.autoRefreshRate * 1000,
  });
  const { data: kubernetesEndpoints, ...kubernetesEndpointsQuery } =
    useKubernetesEndpointsQuery(environmentId, {
      autoRefreshRate: tableState.autoRefreshRate * 1000,
    });
  const { data: environment, ...environmentQuery } =
    useEnvironment(environmentId);
  const environmentUrl = environment?.URL;
  const isServerMetricsEnabled =
    !!environment?.Kubernetes?.Configuration.UseServerMetrics;
  const nodeRowData = useNodeRowData(
    nodes,
    kubernetesEndpoints,
    environmentUrl
  );

  const { isPureAdmin } = useCurrentUser();

  const authorizedToWriteClusterNode = useAuthorizations('K8sClusterNodeW');
  const authorizedToReadMicroK8sStatus = useAuthorizations('K8sResourcePoolsR');

  const { data: credential, ...credentialQuery } = useCloudCredential(
    environment?.CloudProvider?.CredentialID ?? NaN,
    isPureAdmin // if the user is admin
  );

  // currently only microk8s supports deleting nodes
  const canScaleCluster =
    environment?.CloudProvider?.Provider === 'microk8s' &&
    authorizedToWriteClusterNode;
  const canSSH =
    environment?.CloudProvider?.Provider === 'microk8s' &&
    authorizedToWriteClusterNode;

  const canCheckStatus =
    environment?.CloudProvider?.Provider === 'microk8s' &&
    authorizedToReadMicroK8sStatus;

  return (
    <Datatable<IndexOptional<NodeRowData>>
      dataset={nodeRowData ?? []}
      columns={getColumns(isServerMetricsEnabled, canSSH, canCheckStatus)}
      settingsManager={tableState}
      isLoading={
        nodesQuery.isLoading ||
        kubernetesEndpointsQuery.isLoading ||
        environmentQuery.isLoading
      }
      emptyContentLabel="No Nodes found"
      title="Nodes"
      titleIcon={HardDrive}
      getRowId={(row) => row.metadata?.uid ?? ''}
      isRowSelectable={(row) => !row.original.isPublishedNode}
      disableSelect={!authorizedToWriteClusterNode || !canScaleCluster}
      renderTableActions={(selectedRows) =>
        canScaleCluster && (
          <TableActions
            selectedItems={selectedRows}
            environmentId={environmentId}
            nodeRowData={nodeRowData}
          />
        )
      }
      renderTableSettings={() => (
        <TableSettingsMenu>
          <TableSettingsMenuAutoRefresh
            value={tableState.autoRefreshRate}
            onChange={(value) => tableState.setAutoRefreshRate(value)}
          />
        </TableSettingsMenu>
      )}
      description={
        canScaleCluster &&
        !credential &&
        authorizedToWriteClusterNode &&
        credentialQuery.isFetched && (
          <div className="w-full">
            <TextTip color="orange">
              No SSH credentials found for the current cluster.
            </TextTip>
          </div>
        )
      }
    />
  );
}

function TableActions({
  selectedItems,
  environmentId,
  nodeRowData,
}: {
  selectedItems: IndexOptional<Node>[];
  environmentId: EnvironmentId;
  nodeRowData: NodeRowData[];
}) {
  const router = useRouter();
  const { trackEvent } = useAnalytics();
  const removeNodesMutation = useRemoveNodesMutation(environmentId);

  const { data: isProcessing } = useEnvironment(
    environmentId,
    (env) => env?.StatusMessage?.operationStatus === 'processing'
  );

  return (
    <Authorized authorizations="K8sClusterNodeW">
      <Link to="kubernetes.cluster.nodes.new" className="ml-1">
        <Button
          className="btn-wrapper"
          color="secondary"
          icon={Plus}
          disabled={isProcessing}
          data-cy="k8sNodes-addNodesButton"
        >
          Add nodes
        </Button>
      </Link>
      <Button
        className="btn-wrapper"
        color="dangerlight"
        disabled={selectedItems.length === 0 || isProcessing}
        onClick={async () => {
          onRemoveClick(selectedItems);
        }}
        icon={Trash2}
        data-cy="k8sNodes-removeNodeButton"
      >
        Remove
      </Button>
    </Authorized>
  );

  async function onRemoveClick(selectedItems: IndexOptional<Node>[]) {
    const confirmed = await confirmDelete(
      `Removing a node uninstalls MicroK8s from it. During this time, the cluster may become unreachable. Are you sure you want to remove the selected ${pluralize(
        selectedItems.length,
        'node'
      )}?`
    );
    if (confirmed) {
      const nodeIpToDelete = selectedItems.flatMap(
        (item) => getInternalNodeIpAddress(item) || []
      );
      removeNodesMutation.mutate(nodeIpToDelete, {
        onSuccess: () => {
          notifySuccess(
            'Success',
            'Request to remove nodes successfully submitted.'
          );
          router.stateService.reload();
        },
      });

      const masterNodesToRemoveCount = selectedItems.filter(
        (node) => getRole(node) === 'Control plane'
      ).length;
      const workerNodesToRemoveCount = selectedItems.filter(
        (node) => getRole(node) === 'Worker'
      ).length;
      sendAnalytics(masterNodesToRemoveCount, workerNodesToRemoveCount);
    }
  }

  function sendAnalytics(
    masterNodesToRemoveCount: number,
    workerNodesToRemoveCount: number
  ) {
    const currentMasterNodeCount = nodeRowData.filter(
      (node) => getRole(node) === 'Control plane'
    ).length;
    const currentWorkerNodeCount = nodeRowData.filter(
      (node) => getRole(node) === 'Worker'
    ).length;
    trackEvent('scale-down-k8s-cluster', {
      category: 'kubernetes',
      metadata: {
        provider: K8sDistributionType.MICROK8S,
        currentMasterNodeCount,
        currentWorkerNodeCount,
        masterNodesToRemoveCount,
        workerNodesToRemoveCount,
      },
    });
  }
}

/**
 * This function is used to add the isApi property to the node row data.
 */
function useNodeRowData(
  nodes?: Node[],
  kubernetesEndpoints?: Endpoints[],
  environmentUrl?: string
): NodeRowData[] {
  return useMemo<NodeRowData[]>(() => {
    if (!nodes || !kubernetesEndpoints) {
      return [];
    }
    const subsetAddresses = kubernetesEndpoints?.flatMap(
      (endpoint) =>
        endpoint.subsets?.flatMap((subset) => subset.addresses ?? [])
    );
    const nodeRowData = nodes.map((node) => {
      const nodeAddress = getInternalNodeIpAddress(node);
      // if the node address is in the endpoints subset addresses, then it is an api node
      const isApi = subsetAddresses?.some(
        (subsetAddress) => subsetAddress?.ip === nodeAddress
      );
      // if the environment url includes the node address, then it is a published node
      const isPublishedNode =
        !!nodeAddress &&
        !!environmentUrl &&
        environmentUrl?.includes(nodeAddress);
      return {
        ...node,
        isApi,
        isPublishedNode,
        Name: `${node.metadata?.name}${isApi ? 'api' : ''}` ?? '',
      };
    });
    return nodeRowData;
  }, [nodes, kubernetesEndpoints, environmentUrl]);
}
