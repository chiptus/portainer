import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';

import { getDeploymentOptions, getEndpoint } from '../environment.service';
import { Environment, EnvironmentId } from '../types';

import { queryKeys } from './query-keys';

export function useEnvironment<T = Environment | null>(
  id?: EnvironmentId,
  select?: (environment: Environment | null) => T
) {
  return useQuery(
    id ? queryKeys.item(id) : [],
    () => (id ? getEndpoint(id) : null),
    {
      select,
      ...withError('Failed loading environment'),
      staleTime: 50,
      enabled: !!id,
    }
  );
}

export function useEnvironmentDeploymentOptions(id: EnvironmentId) {
  return useQuery(
    [...queryKeys.item(id), 'deploymentOptions'],
    () => getDeploymentOptions(id),
    {
      enabled: !!id,
      ...withError('Failed loading deployment options'),
    }
  );
}
