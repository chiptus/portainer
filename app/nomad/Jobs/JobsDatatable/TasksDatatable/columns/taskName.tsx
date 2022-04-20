import { Column } from 'react-table';

import { Task } from '@/nomad/types';

export const taskName: Column<Task> = {
  Header: 'Task Name',
  accessor: (row) => row.TaskName || '-',
  id: 'taskName',
  disableFilters: true,
  canHide: true,
};
