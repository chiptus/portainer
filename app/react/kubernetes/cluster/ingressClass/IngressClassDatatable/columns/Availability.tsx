import { CellProps, Column } from 'react-table';

import { SwitchField } from '@@/form-components/SwitchField';

import type { IngressControllerClassMap } from '../../types';

function printAvailability(available: boolean): JSX.Element {
  return (
    <SwitchField
      label=""
      checked={available}
      onChange={(checked) => console.log(checked)}
    />
  );
}

export const Availability: Column<IngressControllerClassMap> = {
  Header: 'Availability',
  accessor: 'Availability',
  Cell: ({ row }: CellProps<IngressControllerClassMap>) =>
    printAvailability(row.original.Availability),
  id: 'availability',
  disableFilters: true,
  canHide: true,
  Filter: () => null,
};
