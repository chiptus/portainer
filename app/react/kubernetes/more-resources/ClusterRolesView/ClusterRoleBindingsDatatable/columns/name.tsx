import { Badge } from '@@/Badge';

import { columnHelper } from './helper';

export const name = columnHelper.accessor(
  (row) => {
    if (row.isSystem) {
      return `${row.name} system`;
    }
    return row.name;
  },
  {
    header: 'Role',
    id: 'name',
    cell: ({ row }) => (
      <div className="flex">
        {row.original.name}
        {row.original.isSystem && (
          <Badge type="success" className="ml-2">
            System
          </Badge>
        )}
      </div>
    ),
  }
);
