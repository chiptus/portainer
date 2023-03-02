import { useCurrentStateAndParams } from '@uirouter/react';

import { EnvironmentId } from '@/react/portainer/environments/types';

export function useEnvironmentId(force = true): EnvironmentId {
  const stateAndParams = useCurrentStateAndParams();
  const environmentIdParam = stateAndParams.params.endpointId;
  const environmentIdPath = stateAndParams.params.id;

  const environmentId = environmentIdParam || environmentIdPath;

  if (!environmentId) {
    if (!force) {
      return 0;
    }

    throw new Error('endpointId url param is required');
  }

  return parseInt(environmentId, 10);
}
