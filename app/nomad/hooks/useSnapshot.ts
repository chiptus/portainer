import _ from 'lodash-es';
import { useQuery, useQueryClient } from 'react-query';

import {
  getEndpoint,
  snapshotEndpoint,
} from '@/portainer/environments/environment.service';
import { EnvironmentId } from '@/portainer/environments/types';
import * as notifications from '@/portainer/services/notifications';

export function useSnapshot(environmentId: EnvironmentId) {
  const queryClient = useQueryClient();

  const key = ['environments', environmentId, 'snapshot'];

  function invalidateQuery() {
    queryClient.invalidateQueries(key);
  }

  const query = useQuery(
    key,
    async () => {
      await snapshotEndpoint(environmentId);
      const ret = await getEndpoint(environmentId);
      return _.get(ret, 'Nomad.Snapshots[0]', {});
    },
    {
      refetchOnWindowFocus: false,
      onError: (err) => {
        notifications.error('Failed loading environment', err as Error, '');
      },
    }
  );

  return { query, invalidateQuery };
}
