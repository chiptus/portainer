import { CellContext } from '@tanstack/react-table';
import { Activity, BarChart } from 'lucide-react';

import { Authorized } from '@/react/hooks/useUser';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';

import { Link } from '@@/Link';
import { Icon } from '@@/Icon';

import { NodeRowData } from '../types';
import { NodeShellButton } from '../../../microk8s/NodeShell';
import { getInternalNodeIpAddress } from '../utils';

import { columnHelper } from './helper';

export function getActions(metricsEnabled: boolean, canSSH: boolean) {
  return columnHelper.accessor(() => '', {
    header: 'Actions',
    enableSorting: false,
    cell: (props) => (
      // eslint-disable-next-line react/jsx-props-no-spreading
      <ActionsCell {...props} metricsEnabled={metricsEnabled} canSSH={canSSH} />
    ),
  });
}

function ActionsCell({
  row: { original: node },
  metricsEnabled,
  canSSH,
}: CellContext<NodeRowData, string> & {
  metricsEnabled: boolean;
  canSSH: boolean;
}) {
  const environmentId = useEnvironmentId();
  const nodeName = node.metadata?.name;

  const nodeIp = getInternalNodeIpAddress(node);

  return (
    <div className="flex gap-1.5">
      <Authorized authorizations="K8sClusterNodeR">
        {metricsEnabled && (
          <Link
            title="Stats"
            to="kubernetes.cluster.node.stats"
            params={{ nodeName }}
            className="flex items-center"
          >
            <Icon icon={BarChart} />
          </Link>
        )}

        {nodeIp && node.isApi && (
          <Link
            title="MicroK8s status"
            to="kubernetes.cluster.node.microk8s-status"
            params={{ nodeName }}
            className="flex items-center"
          >
            <Icon icon={Activity} />
          </Link>
        )}

        {nodeIp && canSSH && (
          <NodeShellButton
            windowTitle="SSH Console"
            btnTitle="SSH Console"
            environmentId={environmentId}
            nodeIp={nodeIp}
          />
        )}
      </Authorized>
    </div>
  );
}
