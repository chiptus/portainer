import _ from 'lodash-es';
import clsx from 'clsx';
import { CellProps, Column } from 'react-table';

import { DefaultFilter } from '@/portainer/components/datatables/components/Filter';
import { Task } from '@/nomad/types';

export const taskStatus: Column<Task> = {
  Header: 'Task Status',
  accessor: 'State',
  id: 'status',
  Filter: DefaultFilter,
  canHide: true,
  sortType: 'string',
  Cell: StateCell,
};

function StateCell({ value }: CellProps<Task, string>) {
  const className = getClassName();

  return <span className={clsx('label', className)}>{value}</span>;

  function getClassName() {
    if (['dead'].includes(_.toLower(value))) {
      return 'label-danger';
    }

    if (['pending'].includes(_.toLower(value))) {
      return 'label-warning';
    }

    return 'label-success';
  }
}
