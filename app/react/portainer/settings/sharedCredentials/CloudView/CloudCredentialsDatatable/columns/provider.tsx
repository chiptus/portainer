import { CellProps, Column } from 'react-table';

import { Credential, credentialTitles } from '../../../types';

export const provider: Column<Credential> = {
  Header: 'Provider',
  accessor: (row) => credentialTitles[row.provider],
  id: 'provider',
  Cell: ProviderCell,
  disableFilters: true,
  Filter: () => null,
  canHide: false,
  sortType: 'string',
};

export function ProviderCell({ value: provider }: CellProps<Credential>) {
  return provider;
}
