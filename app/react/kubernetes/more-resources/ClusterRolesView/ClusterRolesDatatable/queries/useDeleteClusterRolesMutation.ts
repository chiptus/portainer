import { useMutation, useQueryClient } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';

import { queryKeys } from './query-keys';

export function useDeleteClusterRolesMutation(environmentId: EnvironmentId) {
  const queryClient = useQueryClient();
  return useMutation(deleteClusterRoles, {
    onSuccess: () =>
      queryClient.invalidateQueries(queryKeys.list(environmentId)),
    ...withError('Unable to delete cluster roles'),
  });
}

export async function deleteClusterRoles({
  environmentId,
  data,
}: {
  environmentId: EnvironmentId;
  data: string[];
}) {
  try {
    return await axios.post(
      `kubernetes/${environmentId}/cluster_roles/delete`,
      data
    );
  } catch (e) {
    throw parseAxiosError(e, `Unable to delete cluster roles`);
  }
}
