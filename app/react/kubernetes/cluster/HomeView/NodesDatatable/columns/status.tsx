import { CellContext } from '@tanstack/react-table';

import { StatusBadge } from '@@/StatusBadge';

import { NodeRowData } from '../types';

import { columnHelper } from './helper';

export const status = columnHelper.accessor((row) => getStatus(row), {
  header: 'Status',
  cell: StatusCell,
});

function StatusCell({
  row: { original: node },
}: CellContext<NodeRowData, string>) {
  const status = getStatus(node);

  const isDeleting =
    node.metadata?.annotations?.['portainer.ip/removing-node'] === 'true';
  if (isDeleting) {
    return <StatusBadge color="warning">Removing</StatusBadge>;
  }

  return (
    <StatusBadge color={status === 'Ready' ? 'success' : 'warning'}>
      {status}
    </StatusBadge>
  );
}

function getStatus(node: NodeRowData) {
  return (
    node.status?.conditions?.find((condition) => condition.status === 'True')
      ?.type ?? 'Not ready'
  );
}
