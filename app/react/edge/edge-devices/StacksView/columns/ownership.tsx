import { CellContext } from '@tanstack/react-table';

import { ownershipIcon } from '@/portainer/filters/filters';
import { ResourceControlOwnership } from '@/react/portainer/access-control/types';

import { Icon } from '@@/Icon';

import { StackInAsyncSnapshot } from '../types';

import { columnHelper } from './helper';

export const ownership = columnHelper.accessor(
  (row) =>
    (row.ResourceControl && row.ResourceControl.Ownership) ||
    ResourceControlOwnership.ADMINISTRATORS,
  {
    header: 'Ownership',
    id: 'ownership',
    cell: OwnershipCell,
  }
);

function OwnershipCell({
  getValue,
}: CellContext<StackInAsyncSnapshot, string>) {
  const value = getValue();

  return (
    <span className="vertical-center">
      <Icon icon={ownershipIcon(value)} />
      {value || ResourceControlOwnership.ADMINISTRATORS}
    </span>
  );
}
