import { Column } from 'react-table';

import { Job } from '@/nomad/types';

export const status: Column<Job> = {
  Header: 'Job Status',
  accessor: (row) => row.Status || '-',
  id: 'statusName',
  disableFilters: true,
  canHide: true,
};
