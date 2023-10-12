import { compact } from 'lodash';
import { useQuery } from 'react-query';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import { isFulfilled } from '@/portainer/helpers/promise-utils';

import { Role } from '../types';

const queryKeys = {
  list: (environmentId: EnvironmentId) =>
    ['environments', environmentId, 'kubernetes', 'roles'] as const,
};

export function useGetRolesQuery(
  environmentId: EnvironmentId,
  namespaces: string[],
  options?: { autoRefreshRate?: number; enabled?: boolean }
) {
  return useQuery(
    queryKeys.list(environmentId),
    async () => {
      const settledRolesPromise = await Promise.allSettled(
        namespaces.map((ns) => getRoles(environmentId, ns))
      );

      return compact(
        settledRolesPromise.filter(isFulfilled).flatMap((i) => i.value)
      );
    },
    {
      ...withError('Unable to get roles'),
      refetchInterval() {
        return options?.autoRefreshRate ?? false;
      },
      enabled: options?.enabled,
    }
  );
}

async function getRoles(environmentId: EnvironmentId, namespace: string) {
  try {
    const { data: roles } = await axios.get<Role[]>(
      `kubernetes/${environmentId}/namespaces/${namespace}/roles`
    );

    return roles;
  } catch (e) {
    throw parseAxiosError(e, 'Unable to get roles');
  }
}
