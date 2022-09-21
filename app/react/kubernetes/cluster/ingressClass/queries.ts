import { useQuery } from 'react-query';

import { useEnvironmentId } from '@/portainer/hooks/useEnvironmentId';

import { getIngressControllerClassMap } from './utils';

export function useIngressControllerClassMap() {
  const environmentId = useEnvironmentId();

  return useQuery(
    ['environment', environmentId, 'ingressControllerClassMap'],
    () => getIngressControllerClassMap(environmentId),
    {
      enabled: environmentId !== undefined,
    }
  );
}
