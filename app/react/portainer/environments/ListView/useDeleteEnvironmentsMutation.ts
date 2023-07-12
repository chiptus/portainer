import { useMutation, useQueryClient } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError } from '@/react-tools/react-query';
import { isFulfilled, isRejected } from '@/portainer/helpers/promise-utils';
import { notifyError, notifySuccess } from '@/portainer/services/notifications';
import { pluralize } from '@/portainer/helpers/strings';

import { buildUrl } from '../environment.service/utils';
import { EnvironmentId } from '../types';

export function useDeleteEnvironmentsMutation() {
  const queryClient = useQueryClient();
  return useMutation(
    async (
      environments: {
        id: EnvironmentId;
        name: string;
        deleteCluster?: boolean;
      }[]
    ) => {
      const deleteEnvResponses = await Promise.allSettled(
        environments.map((env) => deleteEnvironment(env.id, env.deleteCluster))
      );
      const deleteEnvResponsesWithName = deleteEnvResponses.map(
        (response, index) => ({
          ...response,
          name: environments[index].name,
        })
      );
      const successfulDeletions =
        deleteEnvResponsesWithName.filter(isFulfilled);
      const failedDeletions = deleteEnvResponsesWithName.filter(isRejected);

      return { successfulDeletions, failedDeletions };
    },
    {
      ...withError('Unable to delete environment(s)'),
      onSuccess: ({ successfulDeletions, failedDeletions }) => {
        queryClient.invalidateQueries(['environments']);
        // show an error message for each env that failed to delete
        failedDeletions.forEach((deletion) => {
          notifyError(
            `Failed to remove environment`,
            new Error(
              'reason' in deletion ? deletion.reason.message : ''
            ) as Error
          );
        });
        // show one summary message for all successful deletes
        if (successfulDeletions.length) {
          notifySuccess(
            `${pluralize(
              successfulDeletions.length,
              'Environment'
            )} successfully removed`,
            successfulDeletions.map((deletion) => deletion.name).join(', ')
          );
        }
      },
    }
  );
}

async function deleteEnvironment(id: EnvironmentId, deleteCluster?: boolean) {
  try {
    await axios.delete(buildUrl(id), {
      params: { deleteCluster }, // if true, portainer will attempt to delete the cluster. This feature is only available for microk8s cluster environments
    });
  } catch (e) {
    throw parseAxiosError(e as Error, 'Unable to delete environment');
  }
}
