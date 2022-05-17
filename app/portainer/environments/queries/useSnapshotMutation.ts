import { useMutation, useQueryClient } from 'react-query';

import { snapshotEndpoint } from '@/portainer/environments/environment.service';
import { EnvironmentId } from '@/portainer/environments/types';

export function useSnapshotMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    (environmentId: EnvironmentId) => snapshotEndpoint(environmentId),
    {
      onSuccess(_data, environmentId) {
        return queryClient.invalidateQueries(['environments', environmentId]);
      },
    }
  );
}
