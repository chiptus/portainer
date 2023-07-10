import { Volume } from 'docker-types/generated/1.41';

import {
  COMPOSE_STACK_NAME_LABEL,
  SWARM_STACK_NAME_LABEL,
} from '@/react/constants';

import { DockerVolume } from './types';

export function parseViewModel(response: Volume): DockerVolume {
  const stackName =
    (response.Labels &&
      (response.Labels[COMPOSE_STACK_NAME_LABEL] ||
        response.Labels[SWARM_STACK_NAME_LABEL])) ||
    '-';

  return {
    ...response,
    Id: response.Name,
    StackName: stackName,
    Used: false,
  };
}
