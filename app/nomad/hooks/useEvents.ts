import { useQuery, useQueryClient } from 'react-query';
import { useCurrentStateAndParams } from '@uirouter/react';

import { getTaskEvents } from '@/nomad/rest/getTaskEvents';
import * as notifications from '@/portainer/services/notifications';

export function useEvents() {
  const queryClient = useQueryClient();

  const {
    params: {
      endpointId: environmentID,
      allocationID,
      jobID,
      taskName,
      namespace,
    },
  } = useCurrentStateAndParams();

  if (!environmentID) {
    throw new Error('endpointId url param is required');
  }

  const key = [
    'environments',
    environmentID,
    'nomad',
    'events',
    allocationID,
    jobID,
    taskName,
    namespace,
  ];

  function invalidateQuery() {
    return queryClient.invalidateQueries(key);
  }

  const query = useQuery(
    key,
    () =>
      getTaskEvents(environmentID, allocationID, jobID, taskName, namespace),
    {
      refetchOnWindowFocus: false,
      onError: (err) => {
        notifications.error('Failed loading events', err as Error, '');
      },
    }
  );

  return { query, invalidateQuery };
}
