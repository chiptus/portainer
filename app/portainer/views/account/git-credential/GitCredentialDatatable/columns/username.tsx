import { Column } from 'react-table';

import { GitCredential } from '../../types';

export const username: Column<GitCredential> = {
  Header: 'Username',
  accessor: (row) => row.username,
  id: 'username',
  Cell: ({ value }: { value: string }) => value,
  disableFilters: true,
  Filter: () => null,
  canHide: false,
  sortType: 'string',
};
