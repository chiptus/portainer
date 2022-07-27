import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { Environment } from '@/portainer/environments/types';
import { DockerContainerResponse } from '@/react/docker/containers/types/response';
import { parseViewModel } from '@/react/docker/containers/utils';

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
    const { data } = await axios.get<DockerContainerResponse[]>(
      buildDockerSnapshotContainersUrl(environmentId),
      { params }
    );
    return data.map((c) => parseViewModel(c));
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
