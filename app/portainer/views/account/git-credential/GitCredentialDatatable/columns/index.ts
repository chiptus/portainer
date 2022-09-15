import { useMemo } from 'react';

import { creationDate } from './createdDate';
import { name } from './name';
import { username } from './username';

export function useColumns() {
  return useMemo(() => [name, username, creationDate], []);
}
