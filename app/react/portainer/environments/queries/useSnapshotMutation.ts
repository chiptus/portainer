import { useMutation, useQueryClient } from 'react-query';

import { snapshotEndpoint } from '@/react/portainer/environments/environment.service';
import { EnvironmentId } from '@/react/portainer/environments/types';

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
