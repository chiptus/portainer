import { Column } from 'react-table';

import { isoDateFromTimestamp } from '@/portainer/filters/filters';
import { DockerImage } from '@/react/docker/images/types';

export const created: Column<DockerImage> = {
  Header: 'Created',
  accessor: 'Created',
  id: 'created',
  Cell: ({ value }) => isoDateFromTimestamp(value),
  disableFilters: true,
  canHide: true,
  sortType: 'number',
  Filter: () => null,
};
