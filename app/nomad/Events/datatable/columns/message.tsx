import { Column } from 'react-table';

import { NomadEvent } from '@/nomad/types';

export const message: Column<NomadEvent> = {
  Header: 'Message',
  accessor: 'Message',
  id: 'message',
  disableFilters: true,
  canHide: true,
};
