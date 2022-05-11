import { useMemo } from 'react';

import { name } from './name';
import { provider } from './provider';

export function useColumns() {
  return useMemo(() => [name, provider], []);
}
