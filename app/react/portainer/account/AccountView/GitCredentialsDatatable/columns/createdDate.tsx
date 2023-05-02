import { isoDateFromTimestamp } from '@/portainer/filters/filters';

import { columnHelper } from './helper';

export const createdDate = columnHelper.accessor(
  (row) => isoDateFromTimestamp(row.creationDate),
  { id: 'createdDate' }
);
