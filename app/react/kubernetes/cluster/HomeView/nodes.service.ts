import { NodeList } from 'kubernetes-types/core/v1';
import { useMutation, useQuery } from 'react-query';

import axios from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import { queryClient, withError } from '@/react-tools/react-query';

import { AddNodesFormValues } from '../NodeCreateView/types';

// getNodes is used to get a list of nodes using the kubernetes API
async function getNodes(environmentId: EnvironmentId) {
  const { data: nodeList } = await axios.get<NodeList>(
    `/endpoints/${environmentId}/kubernetes/api/v1/nodes`
  );
  return nodeList.items;
}

// useNodesQuery is used to get an array of nodes using the kubernetes API
export function useNodesQuery(
  environmentId: EnvironmentId,
  options?: { autoRefreshRate?: number }
) {
  return useQuery(
    ['environments', environmentId, 'kubernetes', 'nodes'],
    async () => getNodes(environmentId),
    {
      ...withError('Unable to get nodes.'),
      refetchInterval() {
        return options?.autoRefreshRate ?? false;
      },
    }
  );
}

// remove nodes uses the internal portainer API to remove a node from a microk8s cluster
async function removeNodes(
  environmentId: EnvironmentId,
  nodesToRemove: string[]
) {
  await axios.post(`/cloud/endpoints/${environmentId}/nodes/remove`, {
    nodesToRemove,
  });
}

// useRemoveNodesMutation is used to remove a node from a microk8s cluster
export function useRemoveNodesMutation(environmentId: EnvironmentId) {
  return useMutation(
    (nodesToRemove: string[]) => removeNodes(environmentId, nodesToRemove),
    {
      ...withError('Unable to remove nodes.'),
      onSuccess: () =>
        queryClient.invalidateQueries([
          'environments',
          environmentId,
          'kubernetes',
          'nodes',
        ]),
    }
  );
}

async function addNodes(
  environmentId: EnvironmentId,
  addNodesValues: AddNodesFormValues // the formvalues and the request payload match
) {
  await axios.post(
    `/cloud/endpoints/${environmentId}/nodes/add`,
    addNodesValues
  );
}

// useAddNodesMutation is used to nodes to an existing microk8s cluster
export function useAddNodesMutation(environmentId: EnvironmentId) {
  return useMutation(
    (addNodesValues: AddNodesFormValues) =>
      addNodes(environmentId, addNodesValues),
    {
      ...withError('Unable to add nodes.'),
      onSuccess: () =>
        queryClient.invalidateQueries([
          'environments',
          environmentId,
          'kubernetes',
          'nodes',
        ]),
    }
  );
}