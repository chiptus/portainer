import { CellContext } from '@tanstack/react-table';

import { isoDateFromTimestamp } from '@/portainer/filters/filters';
import { useUser } from '@/portainer/users/queries/useUser';

import { EdgeConfiguration } from '../../types';

import { columnHelper } from './helper';

export const updated = columnHelper.accessor('updated', {
  id: 'updated',
  header: 'Updated',
  cell: Cell,
});

function Cell({
  getValue,
  row: {
    original: { createdBy },
  },
}: CellContext<EdgeConfiguration, EdgeConfiguration['updated']>) {
  const value = getValue();
  const { data } = useUser(createdBy);

  let content = '-';
  if (value) {
    content = isoDateFromTimestamp(value);
    if (data) {
      content += ` by ${data.Username}`;
    }
  }

  return <span>{content}</span>;
}
