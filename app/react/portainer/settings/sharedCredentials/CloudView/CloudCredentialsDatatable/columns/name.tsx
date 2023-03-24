import { CellProps, Column } from 'react-table';

import { Link } from '@@/Link';

import { Credential } from '../../../types';

export const name: Column<Credential> = {
  Header: 'Name',
  accessor: (row) => row.name,
  id: 'name',
  Cell: NameCell,
  disableFilters: true,
  Filter: () => null,
  canHide: false,
  sortType: 'string',
};

export function NameCell({ value: name, row }: CellProps<Credential>) {
  return (
    <Link
      to="portainer.settings.sharedcredentials.credential"
      params={{ id: row.id }}
      title={name}
    >
      {name}
    </Link>
  );
}