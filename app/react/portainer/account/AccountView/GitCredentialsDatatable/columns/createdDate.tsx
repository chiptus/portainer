import { Column } from 'react-table';

import { isoDateFromTimestamp } from '@/portainer/filters/filters';
import { GitCredential } from '@/react/portainer/account/git-credentials/types';

export const creationDate: Column<GitCredential> = {
  Header: 'Created',
  accessor: (row) => row.creationDate,
  id: 'creationDate',
  Cell: ({ value }: { value: number }) => isoDateFromTimestamp(value),
  disableFilters: true,
  canHide: true,
  Filter: () => null,
};
