import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { queryKeys } from './query-keys';
import { buildUrl } from './urls';

interface Options<T> {
  select?: (data: Record<EnvironmentId, string>) => T;
  enabled?: boolean;
}

export function usePreviousVersions<T = Record<EnvironmentId, string>>(
  environmentIds: EnvironmentId[],
  { select, enabled }: Options<T> = {}
) {
  return useQuery(
    queryKeys.previousVersions(environmentIds),
    () => getPreviousVersions(environmentIds),
    {
      select,
      enabled: enabled && environmentIds.length > 0,
    }
  );
}

async function getPreviousVersions(environmentIds: EnvironmentId[]) {
  try {
    const { data } = await axios.get<Record<EnvironmentId, string>>(
      buildUrl(undefined, 'previous_versions'),
      {
        params: { environmentIds },
      }
    );
    return data;
  } catch (err) {
    throw parseAxiosError(
      err as Error,
      'Failed to get list of edge update schedules'
    );
  }
}
