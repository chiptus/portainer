import _ from 'lodash';
import { useMemo } from 'react';

import { name } from './name';
import { stackType } from './stackType';
import { ownership } from './ownership';
import { created } from './created';
import { control } from './control';

export function useColumns() {
  return useMemo(
    () => _.compact([name, stackType, ownership, created, control]),
    []
  );
}
