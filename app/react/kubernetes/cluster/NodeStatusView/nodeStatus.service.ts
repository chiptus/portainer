import { useQuery } from 'react-query';

import axios from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import { withError } from '@/react-tools/react-query';

type NodeStatusResponse = {
  status: string;
};

async function getNodeStatus(environmentId: EnvironmentId, nodeIP?: string) {
  if (!nodeIP) {
    return 'Unknown';
  }

  const { data: endpointsList } = await axios.get<NodeStatusResponse>(
    `/cloud/endpoints/${environmentId}/nodes/nodestatus`,
    {
      params: { nodeIP },
    }
  );
  return endpointsList.status;
}

export function useNodeStatusQuery(
  environmentId: EnvironmentId,
  nodeName: string,
  nodeIP?: string
) {
  return useQuery(
    ['environments', environmentId, 'cloud', 'nodes', nodeName, 'nodeStatus'],
    () => getNodeStatus(environmentId, nodeIP),
    {
      ...withError(
        'Unable to retrieve MicroK8s node status',
        'Failed to get node status'
      ),
      enabled: !!nodeIP,
    }
  );
}
