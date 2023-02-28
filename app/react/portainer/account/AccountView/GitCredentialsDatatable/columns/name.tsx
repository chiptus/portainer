import { CellProps, Column } from 'react-table';

import { GitCredential } from '@/react/portainer/account/git-credentials/types';

import { Link } from '@@/Link';

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
