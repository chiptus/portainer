import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { queryKeys } from './query-keys';
import { buildUrl } from './urls';

interface Options<T> {
  select?: (data: Record<EnvironmentId, string>) => T;
  onSuccess?(data: T): void;
  enabled?: boolean;
  skipScheduleID?: number;
}

export function usePreviousVersions<T = Record<EnvironmentId, string>>({
  select,
  onSuccess,
  enabled,
  skipScheduleID = 0,
}: Options<T> = {}) {
  return useQuery(
    queryKeys.previousVersions(skipScheduleID),
    () => getPreviousVersions(skipScheduleID),
    {
      select,
      onSuccess,
      enabled,
    }
  );
}

async function getPreviousVersions(skipScheduleID: number) {
  try {
    const { data } = await axios.get<Record<EnvironmentId, string>>(
      buildUrl(undefined, 'previous_versions'),
      {
        params: {
          skipScheduleID,
        },
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
