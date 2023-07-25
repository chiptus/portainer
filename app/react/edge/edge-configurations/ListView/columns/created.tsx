import { CellContext } from '@tanstack/react-table';

import { isoDateFromTimestamp } from '@/portainer/filters/filters';
import { useUser } from '@/portainer/users/queries/useUser';

import { EdgeConfiguration } from '../../types';

import { columnHelper } from './helper';

export const created = columnHelper.accessor('created', {
  id: 'created',
  header: 'Created',
  cell: Cell,
});

function Cell({
  getValue,
  row: {
    original: { createdBy },
  },
}: CellContext<EdgeConfiguration, EdgeConfiguration['created']>) {
  const value = getValue();
  const { data } = useUser(createdBy);

  let content = isoDateFromTimestamp(value);
  if (data) {
    content += ` by ${data.Username}`;
  }

  return <span>{content}</span>;
}
