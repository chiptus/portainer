import { useCurrentStateAndParams } from '@uirouter/react';

import { EnvironmentId } from '../environments/types';

export function useEnvironmentId(): EnvironmentId {
  const {
    params: { endpointId: environmentId },
  } = useCurrentStateAndParams();

  if (!environmentId) {
    throw new Error('endpointId url param is required');
  }

  return parseInt(environmentId, 10);
}
