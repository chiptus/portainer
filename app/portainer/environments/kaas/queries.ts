import { useMutation, useQuery, useQueryClient } from 'react-query';

import {
  Credential,
  KaasProvider,
  providerTitles,
} from '@/portainer/settings/cloud/types';
import { success as notifySuccess } from '@/portainer/services/notifications';

import { getKaasInfo, createKaasEnvironment } from './kaas.service';
import { CreateClusterPayload } from './types';

export function useCloudProviderOptions(
  credential: Credential,
  provider: KaasProvider
) {
  return useQuery(
    ['cloud', credential.provider, 'info', credential.id],
    () => getKaasInfo(credential.provider, credential),
    {
      meta: {
        error: {
          title: `Failed to get ${providerTitles[credential.provider]} info`,
          message: '',
        },
      },
      enabled: credential.provider === provider,
      retry: 1,
      staleTime: 10000,
    }
  );
}

export function useCreateKaasCluster() {
  const client = useQueryClient();
  return useMutation(
    ({
      payload,
      provider,
    }: {
      payload: CreateClusterPayload;
      provider: KaasProvider;
    }) => createKaasEnvironment(payload, provider),
    {
      onSuccess: (_, { provider }) => {
        notifySuccess('Success', 'KaaS cluster provisioning started');
        client.cancelQueries(['cloud', provider, 'info']);
        return client.invalidateQueries(['environments']);
      },
      meta: {
        error: {
          title: 'Failure',
          message: 'Unable to create KaaS environment',
        },
      },
    }
  );
}
