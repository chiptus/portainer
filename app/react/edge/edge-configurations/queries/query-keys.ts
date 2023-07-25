import { EdgeConfiguration } from '../types';

import { Query as ListQuery } from './list/types';

export const queryKeys = {
  base: () => ['edge', 'configurations'] as const,
  list: (query: ListQuery) => [...queryKeys.base(), query] as const,
  item: (id: EdgeConfiguration['id']) => [...queryKeys.base(), id] as const,
};
