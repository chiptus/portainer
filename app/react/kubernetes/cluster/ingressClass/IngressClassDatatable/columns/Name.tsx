import { CellProps, Column } from 'react-table';

import type { IngressControllerClassMap } from '../../types';

export const Name: Column<IngressControllerClassMap> = {
  Header: 'Ingress class name',
  accessor: 'ClassName',
  Cell: ({ row }: CellProps<IngressControllerClassMap>) =>
    row.original.ClassName || '-',
  id: 'className',
  disableFilters: true,
  canHide: true,
  Filter: () => null,
};
