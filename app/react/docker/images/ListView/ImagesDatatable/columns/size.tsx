import { Column } from 'react-table';

import { DockerImage } from '@/react/docker/images/types';
import { humanize } from '@/portainer/filters/filters';

export const size: Column<DockerImage> = {
  Header: 'Size',
  accessor: 'VirtualSize',
  id: 'size',
  Cell: ({ value }) => humanize(value),
  disableFilters: true,
  canHide: true,
  sortType: 'number',
  Filter: () => null,
};
