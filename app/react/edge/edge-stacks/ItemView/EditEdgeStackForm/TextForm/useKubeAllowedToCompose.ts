import { useCurrentStateAndParams } from '@uirouter/react';
import _ from 'lodash';

import { EnvironmentType } from '@/react/portainer/environments/types';
import { useEdgeGroups } from '@/react/edge/edge-groups/queries/useEdgeGroups';
import { EdgeGroup } from '@/react/edge/edge-groups/types';
import { useEdgeStack } from '@/react/edge/edge-stacks/queries/useEdgeStack';
import { EdgeStack, DeploymentType } from '@/react/edge/edge-stacks/types';

export function useKubeAllowedToCompose() {
  const {
    params: { stackId },
  } = useCurrentStateAndParams();

  const stackQuery = useEdgeStack(stackId);
  const edgeGroupsQuery = useEdgeGroups();

  const stack = stackQuery.data;
  const edgeGroups = edgeGroupsQuery.data;

  if (!stack || !edgeGroups) {
    return false;
  }

  return isKubeAllowToSelectCompose(stack, edgeGroups);
}

function isKubeAllowToSelectCompose(
  stack: EdgeStack,
  edgeGroups: Array<EdgeGroup>
) {
  const stackEdgeGroups = _.compact(
    stack.EdgeGroups.map((id) => edgeGroups.find((e) => e.Id === id))
  );
  const endpointTypes = stackEdgeGroups.flatMap((group) => group.EndpointTypes);
  const initiallyContainsKubeEnv = endpointTypes.includes(
    EnvironmentType.EdgeAgentOnKubernetes
  );
  const isComposeStack = stack.DeploymentType === DeploymentType.Compose;

  return initiallyContainsKubeEnv && isComposeStack;
}
