import { useMemo } from 'react';

import { Availability } from './Availability';
import { Type } from './Type';
import { Name } from './Name';

export function useColumns() {
  return useMemo(() => [Name, Type, Availability], []);
}
