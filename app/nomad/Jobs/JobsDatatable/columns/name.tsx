import { CellProps, Column } from 'react-table';

import { ExpandingCell } from '@/portainer/components/datatables/components/ExpandingCell';
import { Job } from '@/nomad/types';

export const name: Column<Job> = {
  Header: 'Name',
  accessor: (row) => row.ID,
  id: 'name',
  Cell: NameCell,
  disableFilters: true,
  Filter: () => null,
  canHide: false,
  sortType: 'string',
};

export function NameCell({ value: name, row }: CellProps<Job>) {
  return (
    <ExpandingCell row={row} showExpandArrow>
      {name}
    </ExpandingCell>
  );
}
