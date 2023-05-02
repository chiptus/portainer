import { isoDate } from '@/portainer/filters/filters';

import { columnHelper } from './helper';

export const created = columnHelper.accessor('CreatedAt', {
  id: 'created',
  header: 'Created',
  cell: ({ getValue }) => {
    const value = getValue();
    return isoDate(value);
  },
});
