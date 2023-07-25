import { sortOptionsFromColumns } from '@/react/common/api/sort.types';

import { buildNameColumn } from '@@/datatables/NameCell';

import { EdgeConfiguration } from '../../types';

import { created } from './created';
import { groups } from './groups';
import { progress } from './progress';
import { updated } from './updated';

export const columns = [
  buildNameColumn<EdgeConfiguration>('name', 'id', '.item'),
  groups,
  created,
  updated,
  progress,
];

export const sortOptions = sortOptionsFromColumns(columns);
