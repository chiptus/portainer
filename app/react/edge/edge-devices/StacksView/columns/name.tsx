import _ from 'lodash';
import { CellContext } from '@tanstack/react-table';

import { TableSettings } from '@/react/docker/containers/ListView/ContainersDatatable/types';

import { useTableSettings } from '@@/datatables/useTableSettings';

import { StackInAsyncSnapshot } from '../types';

import { columnHelper } from './helper';

export const name = columnHelper.accessor((row) => row.Names[0], {
  header: 'Name',
  id: 'name',
  cell: NameCell,
});

function NameCell({ getValue }: CellContext<StackInAsyncSnapshot, string>) {
  const name = getValue();

  const settings = useTableSettings<TableSettings>();
  const truncate = settings.truncateContainerName;

  let shortName = name;
  if (truncate > 0) {
    shortName = _.truncate(name, { length: truncate });
  }

  return shortName;
}
