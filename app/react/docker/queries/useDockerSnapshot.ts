import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { Environment } from '@/react/portainer/environments/types';
import { DockerSnapshotResponse } from '@/react/docker/snapshots/types/response';
import { parseViewModel } from '@/react/docker/snapshots/utils';

import { buildDockerSnapshotUrl, queryKeys } from './utils';

export function useDockerSnapshot(environmentId: Environment['Id']) {
  return useQuery(
    queryKeys.snapshotQuery(environmentId),
    () => getEnvironmentSnapshot(environmentId),
    {
      ...withError('Failed loading snapshot'),
    }
  );
}

export async function getEnvironmentSnapshot(environmentId: Environment['Id']) {
  try {
    const { data } = await axios.get<DockerSnapshotResponse>(
      buildDockerSnapshotUrl(environmentId)
    );
    return parseViewModel(data);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}