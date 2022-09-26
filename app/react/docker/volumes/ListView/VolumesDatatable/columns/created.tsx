import { Column } from 'react-table';

import { isoDate } from '@/portainer/filters/filters';
import { DockerVolume } from '@/react/docker/volumes/types';

export const created: Column<DockerVolume> = {
  Header: 'Created',
  accessor: 'CreatedAt',
  id: 'created',
  Cell: ({ value }) => (value ? isoDate(value) : '-'),
  disableFilters: true,
  canHide: true,
  sortType: 'string',
  Filter: () => null,
};
