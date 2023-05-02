import { truncateLeftRight } from '@/portainer/filters/filters';

import { columnHelper } from './helper';

export const mountPoint = columnHelper.accessor('Mountpoint', {
  header: 'Mount point',
  id: 'mountPoint',
  cell: ({ getValue }) => {
    const value = getValue();

    return truncateLeftRight(value);
  },
});
