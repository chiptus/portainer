import { DockerVolume } from './types';
import { DockerVolumeResponse } from './types/response';

export function parseViewModel(response: DockerVolumeResponse): DockerVolume {
  const stackName =
    (response.Labels &&
      (response.Labels['com.docker.compose.project'] ||
        response.Labels['com.docker.stack.namespace'])) ||
    '-';

  return {
    ...response,
    Id: response.Name,
    StackName: stackName,
    Used: false,
  };
}
