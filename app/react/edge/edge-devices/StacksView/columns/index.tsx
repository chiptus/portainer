import _ from 'lodash';
import { useMemo } from 'react';

import { createOwnershipColumn } from '@/react/docker/components/datatables/createOwnershipColumn';

import { StackInAsyncSnapshot } from '../types';

import { name } from './name';
import { stackType } from './stackType';
import { created } from './created';
import { control } from './control';

export function useColumns() {
  return useMemo(
    () =>
      _.compact([
        name,
        stackType,
        createOwnershipColumn<StackInAsyncSnapshot>(),
        created,
        control,
      ]),
    []
  );
}
