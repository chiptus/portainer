import { columnHelper } from './helper';

export const roleKind = columnHelper.accessor('roleRef.kind', {
  header: 'Kind',
  id: 'kind',
});
