import { useQuery } from 'react-query';
import { compact } from 'lodash';

import { withError } from '@/react-tools/react-query';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { EnvironmentId } from '@/react/portainer/environments/types';
import { isFulfilled } from '@/portainer/helpers/promise-utils';

import { ServiceAccount } from '../../types';

import { queryKeys } from './query-keys';

export function useGetServiceAccountsQuery(
  environmentId: EnvironmentId,
  namespaces: string[],
  options?: {
    autoRefreshRate?: number;
    enabled?: boolean;
  }
) {
  return useQuery(
    queryKeys.list(environmentId),
    async () => {
      const settledServicesPromise = await Promise.allSettled(
        namespaces.map((namespace) =>
          getServiceAccounts(environmentId, namespace)
        )
      );
      return compact(
        settledServicesPromise.filter(isFulfilled).flatMap((i) => i.value)
      );
    },
    {
      ...withError('Unable to get service accounts'),
      refetchInterval() {
        return options?.autoRefreshRate ?? false;
      },
      enabled: options?.enabled,
    }
  );
}

async function getServiceAccounts(
  environmentId: EnvironmentId,
  namespace: string
) {
  try {
    const { data: services } = await axios.get<ServiceAccount[]>(
      `kubernetes/${environmentId}/namespaces/${namespace}/service_accounts`
    );

    return services;
  } catch (e) {
    throw parseAxiosError(e, 'Unable to get service accounts');
  }
}
