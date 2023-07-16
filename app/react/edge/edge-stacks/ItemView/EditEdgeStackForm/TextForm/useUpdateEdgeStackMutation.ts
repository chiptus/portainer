import { useMutation, useQueryClient } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import {
  mutationOptions,
  withError,
  withInvalidate,
} from '@/react-tools/react-query';
import { buildUrl } from '@/react/edge/edge-stacks/queries/buildUrl';
import { DeploymentType, EdgeStack } from '@/react/edge/edge-stacks/types';
import { EdgeGroup } from '@/react/edge/edge-groups/types';
import { queryKeys } from '@/react/edge/edge-stacks/queries/query-keys';
import { RegistryId } from '@/react/portainer/registries/types';

import { Value as EnvVars } from '@@/form-components/EnvironmentVariablesFieldset/types';

interface UpdateEdgeStackPayload {
  id: EdgeStack['Id'];
  stackFileContent: string;
  edgeGroups: Array<EdgeGroup['Id']>;
  deploymentType: DeploymentType;
  registries: Array<RegistryId>;
  /** Uses the manifest's namespaces instead of the default one */
  useManifestNamespaces: boolean;
  prePullImage: boolean;
  rePullImage: boolean;
  retryDeploy: boolean;
  updateVersion: boolean;
  /** Optional webhook configuration */
  webhook?: string;
  /** Environment variables to inject into the stack */
  envVars: EnvVars;
  /** RollbackTo specifies the stack file version to rollback to */
  rollbackTo?: number;
}

export function useUpdateEdgeStackMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    updateEdgeStack,
    mutationOptions(
      withError('Failed updating stack'),
      withInvalidate(queryClient, [queryKeys.base()])
    )
  );
}

async function updateEdgeStack({ id, ...payload }: UpdateEdgeStackPayload) {
  try {
    await axios.put(buildUrl(id), payload);
  } catch (err) {
    throw parseAxiosError(err as Error, 'Failed updating stack');
  }
}
