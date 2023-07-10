import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { ContainerId } from '@/react/docker/containers/types';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { urlBuilder } from '../containers.service';
import { DockerContainerResponse } from '../types/response';
import { parseViewModel } from '../utils';

import { queryKeys } from './query-keys';

export function useContainer(
  environmentId: EnvironmentId,
  containerId: ContainerId
) {
  return useQuery(
    queryKeys.container(environmentId, containerId),
    () => getContainer(environmentId, containerId),
    {
      meta: {
        title: 'Failure',
        message: 'Unable to retrieve container',
      },
    }
  );
}

async function getContainer(
  environmentId: EnvironmentId,
  containerId: ContainerId
) {
  try {
    const { data } = await axios.get<DockerContainerResponse>(
      urlBuilder(environmentId, containerId, 'json')
    );
    return parseViewModel(data);
  } catch (error) {
    throw parseAxiosError(error as Error, 'Unable to retrieve container');
  }
}
