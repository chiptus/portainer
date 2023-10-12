import { compact } from 'lodash';
import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import { isFulfilled } from '@/portainer/helpers/promise-utils';

import { RoleBinding } from '../types';

import { queryKeys } from './query-keys';

export function useRoleBindingsQuery(
  environmentId: EnvironmentId,
  namespaces: string[],
  options?: { autoRefreshRate?: number; enabled?: boolean }
) {
  return useQuery(
    queryKeys.list(environmentId),
    async () => {
      const settledServicesPromise = await Promise.allSettled(
        namespaces.map((namespace) => getRoleBindings(environmentId, namespace))
      );
      return compact(
        settledServicesPromise.filter(isFulfilled).flatMap((i) => i.value)
      );
    },
    {
      ...withError('Unable to get role bindings'),
      refetchInterval() {
        return options?.autoRefreshRate ?? false;
      },
      enabled: options?.enabled,
    }
  );
}

async function getRoleBindings(
  environmentId: EnvironmentId,
  namespace: string
) {
  try {
    const { data: roleBinding } = await axios.get<RoleBinding[]>(
      `kubernetes/${environmentId}/namespaces/${namespace}/role_bindings`
    );

    return roleBinding;
  } catch (e) {
    throw parseAxiosError(e, 'Unable to get role bindings');
  }
}
