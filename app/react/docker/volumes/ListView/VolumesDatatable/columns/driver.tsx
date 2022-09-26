import { Column } from 'react-table';

import { DockerVolume } from '@/react/docker/volumes/types';

export const driver: Column<DockerVolume> = {
  Header: 'Driver',
  accessor: 'Driver',
  id: 'driver',
  disableFilters: true,
  canHide: true,
  sortType: 'string',
  Filter: () => null,
};
