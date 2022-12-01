import { useQuery } from 'react-query';

import { ServiceId } from '@/react/docker/services/types';
import { getService } from '@/react/docker/services/axios/getService';
import { queryKeys } from '@/react/docker/services/queries/query-keys';
import { EnvironmentId } from '@/react/portainer/environments/types';

export function useService(environmentId: EnvironmentId, serviceId: ServiceId) {
  return useQuery(
    queryKeys.service(environmentId, serviceId),
    () => getService(environmentId, serviceId),
    {
      enabled: !!serviceId,
      meta: {
        title: 'Failure',
        message: 'Unable to retrieve service',
      },
    }
  );
}
