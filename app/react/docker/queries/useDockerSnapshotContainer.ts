import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { Environment } from '@/react/portainer/environments/types';

import { DockerContainerSnapshotResponse } from '../snapshots/types/response';
import { parseDockerContainerSnapshot } from '../snapshots/utils';

import { buildDockerSnapshotContainersUrl, queryKeys } from './utils';

export function useDockerSnapshotContainer(
  environmentId: Environment['Id'],
  containerId: DockerContainerSnapshotResponse['Id']
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
  containerId: DockerContainerSnapshotResponse['Id']
) {
  try {
    const { data } = await axios.get<DockerContainerSnapshotResponse>(
      buildDockerSnapshotContainersUrl(environmentId, containerId)
    );
    return parseDockerContainerSnapshot(data);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
