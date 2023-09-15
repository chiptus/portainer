import { CellContext } from '@tanstack/react-table';
import { Activity, BarChart, Terminal } from 'lucide-react';
import { v4 as uuidv4 } from 'uuid';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { getBaseUrl } from '@/portainer/helpers/webhookHelper';
import { useAnalytics } from '@/react/hooks/useAnalytics';

import { Link } from '@@/Link';
import { Icon } from '@@/Icon';
import { Button } from '@@/buttons';

import { NodeRowData } from '../types';
import { getInternalNodeIpAddress, getRole } from '../utils';

import { columnHelper } from './helper';

export function getActions(
  metricsEnabled: boolean,
  canSSH: boolean,
  canCheckStatus: boolean
) {
  return columnHelper.accessor(() => '', {
    header: 'Actions',
    enableSorting: false,
    cell: (props) => (
      <ActionsCell
        // eslint-disable-next-line react/jsx-props-no-spreading
        {...props}
        metricsEnabled={metricsEnabled}
        canSSH={canSSH}
        canCheckStatus={canCheckStatus}
      />
    ),
  });
}

function ActionsCell({
  row: { original: node },
  metricsEnabled,
  canSSH,
  canCheckStatus,
}: CellContext<NodeRowData, string> & {
  metricsEnabled: boolean;
  canSSH: boolean;
  canCheckStatus: boolean;
}) {
  const { trackEvent } = useAnalytics();
  const url = getBaseUrl();
  const environmentId = useEnvironmentId();
  const nodeName = node.metadata?.name;

  const nodeIp = getInternalNodeIpAddress(node);
  const nodeRole = getRole(node);

  return (
    <div className="flex gap-1.5">
      {metricsEnabled && (
        <Link
          title="Stats"
          to="kubernetes.cluster.node.stats"
          params={{ nodeName }}
          className="flex items-center p-1"
        >
          <Icon icon={BarChart} />
        </Link>
      )}
      {nodeIp && nodeRole === 'Control plane' && canCheckStatus && (
        <Link
          title="MicroK8s status"
          to="kubernetes.cluster.node.microk8s-status"
          params={{ nodeName }}
          className="flex items-center p-1"
        >
          <Icon icon={Activity} />
        </Link>
      )}

      {nodeIp && nodeRole === 'Worker' && canCheckStatus && (
        <div
          className="text-muted flex items-center p-1"
          title="Status - not available for worker nodes"
        >
          <Icon icon={Activity} />
        </div>
      )}

      {nodeIp && canSSH && (
        <Button
          title="SSH Console"
          color="none"
          size="small"
          data-cy="nodeShellButton"
          className="!ml-0 !p-1 !text-blue-8"
          icon={Terminal}
          onClick={() => {
            window.open(
              `${url}#!/${environmentId}/kubernetes/node-shell?nodeIP=${nodeIp}`,
              // give the window a unique name so that more than one can be opened
              `node-shell-${nodeName}-${uuidv4()}`,
              'width=800,height=600'
            );
            trackEvent('microk8s-shell', { category: 'kubernetes' });
          }}
        />
      )}
    </div>
  );
}
