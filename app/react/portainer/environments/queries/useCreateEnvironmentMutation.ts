import { useQueryClient, useMutation, MutationFunction } from 'react-query';

import {
  mutationOptions,
  withError,
  withInvalidate,
} from '@/react-tools/react-query';

import {
  createRemoteEnvironment,
  createLocalDockerEnvironment,
  createAzureEnvironment,
  createAgentEnvironment,
  createEdgeAgentEnvironment,
  createLocalKubernetesEnvironment,
  createKubeConfigEnvironment,
} from '../environment.service/create';
import { queryKey as nodesCountQueryKey } from '../../system/useNodesCount';

export function useCreateAzureEnvironmentMutation() {
  return useGenericCreationMutation(createAzureEnvironment);
}

export function useCreateLocalDockerEnvironmentMutation() {
  return useGenericCreationMutation(createLocalDockerEnvironment);
}

export function useCreateLocalKubernetesEnvironmentMutation() {
  return useGenericCreationMutation(createLocalKubernetesEnvironment);
}

export function useCreateKubeConfigEnvironmentMutation() {
  return useGenericCreationMutation(createKubeConfigEnvironment);
}

export function useCreateRemoteEnvironmentMutation(
  creationType: Parameters<typeof createRemoteEnvironment>[0]['creationType']
) {
  return useGenericCreationMutation(
    (
      params: Omit<
        Parameters<typeof createRemoteEnvironment>[0],
        'creationType'
      >
    ) => createRemoteEnvironment({ creationType, ...params })
  );
}

export function useCreateAgentEnvironmentMutation() {
  return useGenericCreationMutation(createAgentEnvironment);
}

export function useCreateEdgeAgentEnvironmentMutation() {
  return useGenericCreationMutation(createEdgeAgentEnvironment);
}

function useGenericCreationMutation<TData = unknown, TVariables = void>(
  mutation: MutationFunction<TData, TVariables>
) {
  const queryClient = useQueryClient();

  return useMutation(
    mutation,
    mutationOptions(
      withError('Unable to create environment'),
      withInvalidate(queryClient, [['environments'], nodesCountQueryKey])
    )
  );
}
