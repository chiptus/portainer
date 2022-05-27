import { CellProps, Column } from 'react-table';

import { Credential, providerTitles } from '../../types';

export const provider: Column<Credential> = {
  Header: 'Cloud Provider',
  accessor: (row) => providerTitles[row.provider],
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
