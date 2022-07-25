import { useMutation, useQueryClient } from 'react-query';

import { EnvironmentId } from '@/portainer/environments/types';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError } from '@/react-tools/react-query';

import { EdgeStack } from '../types';

import { queryKeys } from './query-keys';

export function useCollectLogsMutation() {
  const queryClient = useQueryClient();

  return useMutation(collectLogs, {
    onSuccess(data, variables) {
      return queryClient.invalidateQueries(
        queryKeys.logsStatus(variables.edgeStackId, variables.environmentId)
      );
    },
    ...withError('Unable to retrieve logs'),
  });
}

interface CollectLogs {
  edgeStackId: EdgeStack['Id'];
  environmentId: EnvironmentId;
}

async function collectLogs({ edgeStackId, environmentId }: CollectLogs) {
  try {
    await axios.put(`/edge_stacks/${edgeStackId}/logs/${environmentId}`);
  } catch (error) {
    throw parseAxiosError(error as Error, 'Unable to start logs collection');
  }
}
