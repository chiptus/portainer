import { Column } from 'react-table';

import { DockerVolume } from '@/react/docker/volumes/types';

export const stackName: Column<DockerVolume> = {
  Header: 'Stack',
  accessor: 'StackName',
  id: 'stackname',
  disableFilters: true,
  canHide: true,
  sortType: 'string',
  Filter: () => null,
};
