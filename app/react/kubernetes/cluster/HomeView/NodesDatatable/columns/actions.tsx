import { CellContext } from '@tanstack/react-table';
import { BarChart } from 'lucide-react';

import { Authorized } from '@/react/hooks/useUser';

import { Link } from '@@/Link';
import { Icon } from '@@/Icon';

import { NodeRowData } from '../types';

import { columnHelper } from './helper';

export const actions = columnHelper.accessor(() => '', {
  header: 'Actions',
  cell: ActionsCell,
});

function ActionsCell({
  row: { original: node },
}: CellContext<NodeRowData, string>) {
  const name = node.metadata?.name;
  return (
    <div className="flex gap-1">
      <Authorized authorizations="K8sClusterNodeR">
        <Link
          to="kubernetes.cluster.node.stats"
          params={{ name }}
          className="flex items-center gap-1"
        >
          <Icon icon={BarChart} />
          Stats
        </Link>
      </Authorized>
    </div>
  );
}
