import { columnHelper } from './helper';

export const subjects = columnHelper.display({
  header: 'Subjects:',
  id: 'subjects',
  cell: '',
  size: 0,
  maxSize: 0,
  meta: {
    className: '[&>div]:block !text-right !pr-0',
  },
});
