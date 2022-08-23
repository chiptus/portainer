import { Column } from 'react-table';

import { DockerContainer } from '@/react/docker/containers/types';
import { trimSHA } from '@/docker/filters/utils';

export const image: Column<DockerContainer> = {
  Header: 'Image',
  accessor: (row) => trimSHA(row.Image),
  id: 'image',
  disableFilters: true,
  canHide: true,
  sortType: 'string',
  Filter: () => null,
};
