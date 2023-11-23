import { createColumnHelper } from '@tanstack/react-table';

import { TableEntry } from '../types';

const columnHelper = createColumnHelper<TableEntry>();

export const columns = [
  columnHelper.accessor('name', {
    header: 'Name',
    id: 'name',
  }),
  columnHelper.accessor('type', {
    header: 'Type',
    id: 'type',
  }),
];
