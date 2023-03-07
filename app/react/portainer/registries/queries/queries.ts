import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';

import { getRegistries } from '../registry.service';
import { Registry } from '../types/registry';

export function useRegistries<T = Registry[]>({
  enabled,
  select,
  onSuccess,
}: {
  enabled?: boolean;
  select?: (registries: Registry[]) => T;
  onSuccess?: (data: T) => void;
} = {}) {
  return useQuery(['registries'], getRegistries, {
    select,
    ...withError('Unable to retrieve registries'),
    enabled,
    onSuccess,
  });
}
