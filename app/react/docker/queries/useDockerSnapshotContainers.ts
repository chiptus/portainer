import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { Environment } from '@/react/portainer/environments/types';
import { DockerContainerResponse } from '@/react/docker/containers/types/response';
import { toListViewModel } from '@/react/docker/containers/utils';

import {
  buildDockerSnapshotContainersUrl,
  ContainersQueryParams,
  queryKeys,
} from './utils';

export function useDockerSnapshotContainers(
  environmentId: Environment['Id'],
  params: ContainersQueryParams
) {
  return useQuery(
    queryKeys.containersQuery(environmentId, params),
    () => getEnvironmentSnapshotContainers(environmentId, params),
    {
      ...withError('Failed loading snapshot containers'),
    }
  );
}

export async function getEnvironmentSnapshotContainers(
  environmentId: Environment['Id'],
  params: ContainersQueryParams
) {
  try {
    const { data } = await axios.get<null | DockerContainerResponse[]>(
      buildDockerSnapshotContainersUrl(environmentId),
      { params }
    );
    return data && data.map((c) => toListViewModel(c));
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
