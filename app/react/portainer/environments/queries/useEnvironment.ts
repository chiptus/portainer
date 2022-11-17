import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';

import { getDeploymentOptions, getEndpoint } from '../environment.service';
import { Environment, EnvironmentId } from '../types';

export function useEnvironment<T = Environment | null>(
  id?: EnvironmentId,
  select?: (environment: Environment | null) => T
) {
  return useQuery(['environments', id], () => (id ? getEndpoint(id) : null), {
    select,
    ...withError('Failed loading environment'),
    staleTime: 50,
    enabled: !!id,
  });
}

export function useEnvironmentDeploymentOptions(id: EnvironmentId) {
  return useQuery(
    ['environments', id, 'deploymentOptions'],
    () => getDeploymentOptions(id),
    {
      enabled: !!id,
      ...withError('Failed loading deployment options'),
    }
  );
}
