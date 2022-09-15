import { CellProps, Column } from 'react-table';

import { Link } from '@@/Link';

import { GitCredential } from '../../types';

export const name: Column<GitCredential> = {
  Header: 'Name',
  accessor: (row) => row.name,
  id: 'name',
  Cell: NameCell,
  disableFilters: true,
  Filter: () => null,
  canHide: false,
  sortType: 'string',
};

export function NameCell({ value: name, row }: CellProps<GitCredential>) {
  return (
    <Link
      to="portainer.account.gitEditGitCredential"
      params={{ id: row.id }}
      title={name}
    >
      {name}
    </Link>
  );
}
