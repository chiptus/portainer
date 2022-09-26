import { Column } from 'react-table';

import { truncateLeftRight } from '@/portainer/filters/filters';
import { DockerVolume } from '@/react/docker/volumes/types';

export const mountpoint: Column<DockerVolume> = {
  Header: 'Mount point',
  accessor: 'Mountpoint',
  id: 'mountpoint',
  Cell: ({ value }) => truncateLeftRight(value),
  disableFilters: true,
  canHide: true,
  sortType: 'string',
  Filter: () => null,
};
