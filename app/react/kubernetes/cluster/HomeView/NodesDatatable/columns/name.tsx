import { CellContext } from '@tanstack/react-table';

import { Authorized } from '@/react/hooks/useUser';

import { Link } from '@@/Link';
import { Badge } from '@@/Badge';

import { NodeRowData } from '../types';

import { columnHelper } from './helper';

export const name = columnHelper.accessor('name', {
  header: 'Name',
  cell: NameCell,
});

function NameCell({
  row: { original: node },
}: CellContext<NodeRowData, string>) {
  const name = node.metadata?.name;
  return (
    <div className="flex gap-2">
      <Authorized authorizations="K8sClusterNodeR" childrenUnauthorized={name}>
        <Link to="kubernetes.cluster.node" params={{ name }}>
          {name}
        </Link>
      </Authorized>
      {node.isPublishedNode && <Badge type="success">connected</Badge>}
      {node.isApi && <Badge type="info">api</Badge>}
    </div>
  );
}
