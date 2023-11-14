import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';

import { EdgeStack } from '../types';

import { buildUrl } from './buildUrl';
import { queryKeys } from './query-keys';

export function useEdgeStacks<T = Array<EdgeStack>>({
  select,
  /**
   * If set to a number, the query will continuously refetch at this frequency in milliseconds.
   * If set to a function, the function will be executed with the latest data and query to compute a frequency
   * Defaults to `false`.
   */
  refetchInterval,
  edgeComputeEnabled = true,
}: {
  select?: (stacks: EdgeStack[]) => T;
  refetchInterval?: number | false | ((data?: T) => false | number);
  edgeComputeEnabled?: boolean;
} = {}) {
  return useQuery(queryKeys.base(), () => getEdgeStacks(edgeComputeEnabled), {
    ...withError('Failed loading Edge stack'),
    select,
    refetchInterval,
  });
}

export async function getEdgeStacks(edgeComputeEnabled: boolean) {
  try {
    if (edgeComputeEnabled) {
      const { data } = await axios.get<EdgeStack[]>(buildUrl());
      return data;
    }

    return [];
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
