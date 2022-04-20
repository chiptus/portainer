import { useMutation, useQuery, useQueryClient } from 'react-query';

import {
  getKaasInfo,
  createKaasEnvironment,
} from '@/portainer/environments/environment.service/kaas';
import {
  error as notifyError,
  success as notifySuccess,
} from '@/portainer/services/notifications';

import { KaasProvider, KaasCreateFormValues } from './kaas.types';

export function useCloudProviderOptions(
  provider: KaasProvider | undefined,
  enabled: boolean
) {
  return useQuery(['cloud', provider, 'info'], () => getKaasInfo(provider), {
    onError: (err) => {
      notifyError(
        'Failure',
        err as Error,
        `Unable to retrieve ${provider} info`
      );
    },
    enabled,
    retry: false,
    staleTime: 10000,
  });
}

export function useCreateKaasCluster() {
  const client = useQueryClient();
  return useMutation(
    (formValues: KaasCreateFormValues) => createKaasEnvironment(formValues),
    {
      onSuccess: () => {
        notifySuccess('Success', 'Environment created');
        return client.invalidateQueries(['environments']);
      },
      onError: (err) => {
        notifyError(
          'Failure',
          err as Error,
          'Unable to create KaaS environment'
        );
      },
    }
  );
}
