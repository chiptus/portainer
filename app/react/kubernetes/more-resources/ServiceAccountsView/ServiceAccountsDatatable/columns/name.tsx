import { Badge } from '@@/Badge';

import { columnHelper } from './helper';

export const name = columnHelper.accessor(
  (row) => {
    let result = row.name;
    if (row.isSystem) {
      result += ' system';
    }
    if (row.isUnused) {
      result += ' unused';
    }
    return result;
  },
  {
    header: 'Name',
    id: 'Name',
    cell: ({ row }) => (
      <div className="flex">
        <div>{row.original.name}</div>
        {row.original.isSystem && (
          <Badge type="success" className="ml-2">
            System
          </Badge>
        )}
        {row.original.isUnused && (
          <Badge type="warn" className="ml-2">
            Unused
          </Badge>
        )}
      </div>
    ),
  }
);
