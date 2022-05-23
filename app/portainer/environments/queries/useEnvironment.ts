import { useQuery } from 'react-query';

import { getEndpoint } from '@/portainer/environments/environment.service';
import { EnvironmentId } from '@/portainer/environments/types';
import { withError } from '@/react-tools/react-query';

export function useEnvironment(id: EnvironmentId) {
  return useQuery(['environments', id], () => getEndpoint(id), {
    ...withError('Failed loading environment'),
    staleTime: 50,
  });
}
