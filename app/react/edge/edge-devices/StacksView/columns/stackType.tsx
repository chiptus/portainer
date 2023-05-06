import { CellContext } from '@tanstack/react-table';

import {
  COMPOSE_STACK_NAME_LABEL,
  SWARM_STACK_NAME_LABEL,
} from '@/react/constants';

import { StackInAsyncSnapshot } from '../types';

import { columnHelper } from './helper';

export const stackType = columnHelper.accessor((row) => row.StackName, {
  header: 'Type',
  id: 'type',
  cell: TypeCell,
});

function TypeCell({
  row: { original: container },
}: CellContext<StackInAsyncSnapshot, string>) {
  const isCompose =
    container.Labels && container.Labels[COMPOSE_STACK_NAME_LABEL];
  const isSwarm = container.Labels && container.Labels[SWARM_STACK_NAME_LABEL];

  if (isCompose) {
    return 'Compose';
  }

  if (isSwarm) {
    return 'Swarm';
  }

  return '';
}
