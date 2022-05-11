import { CellProps, Column } from 'react-table';

import { Credential } from '../../types';

const providerTitles = {
  civo: 'Civo',
  linode: 'Linode',
  digitalocean: 'DigitalOcean',
  googlecloud: 'Google Cloud',
  aws: 'AWS',
  azure: 'Azure',
};

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
