import { useQuery } from 'react-query';

import { EnvironmentId } from '@/portainer/environments/types';
import { Dashboard } from '@/nomad/types';
import { getDashboard } from '@/nomad/rest/getDashboard';

export function useDashboard(environmentId: EnvironmentId) {
  return useQuery<Dashboard>(
    ['environments', environmentId, 'nomad', 'dashboard'],
    () => getDashboard(environmentId),
    {
      meta: {
        error: {
          title: 'Failure',
          message: 'Unable to get dashboard information',
        },
      },
    }
  );
}
