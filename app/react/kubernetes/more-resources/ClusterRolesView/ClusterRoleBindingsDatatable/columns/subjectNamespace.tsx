import { Link } from '@@/Link';

import { columnHelper } from './helper';

export const subjectNamespace = columnHelper.accessor(
  (row) => row.subjects?.map((sub) => sub.namespace).join(' '),
  {
    header: 'Subject Namespace',
    id: 'subjectNamespace',
    cell: ({ row }) =>
      row.original.subjects?.map((sub, index) => (
        <div key={index}>
          {sub.namespace ? (
            <Link
              to="kubernetes.resourcePools.resourcePool"
              params={{
                id: sub.namespace,
              }}
              title={sub.namespace}
            >
              {sub.namespace}
            </Link>
          ) : (
            '-'
          )}
        </div>
      )) || '-',
  }
);
