import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { Environment } from '@/portainer/environments/types';
import { DockerContainerResponse } from '@/react/docker/containers/types/response';
import { parseViewModel } from '@/react/docker/containers/utils';

import { buildDockerSnapshotContainersUrl, queryKeys } from './utils';

export function useDockerSnapshotContainer(
  environmentId: Environment['Id'],
  containerId: DockerContainerResponse['Id']
) {
  return useQuery(
    queryKeys.container(environmentId, containerId),
    () => getEnvironmentSnapshotContainers(environmentId, containerId),
    {
      ...withError('Failed loading snapshot containers'),
    }
  );
}

export async function getEnvironmentSnapshotContainers(
  environmentId: Environment['Id'],
  containerId: DockerContainerResponse['Id']
) {
  try {
    const { data } = await axios.get<DockerContainerResponse>(
      buildDockerSnapshotContainersUrl(environmentId, containerId)
    );
    return parseViewModel(data);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
