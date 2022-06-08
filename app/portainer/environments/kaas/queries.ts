import { useMutation, useQuery, useQueryClient } from 'react-query';

import {
  Credential,
  KaasProvider,
  providerTitles,
} from '@/portainer/settings/cloud/types';
import { success as notifySuccess } from '@/portainer/services/notifications';

import { getKaasInfo, createKaasEnvironment } from './kaas.service';
import { CreateClusterPayload, KaasInfo } from './types';

export function useCloudProviderOptions<T extends KaasInfo>(
  provider: KaasProvider,
  validator: (info: KaasInfo) => info is T,
  credential?: Credential | null,
  force = false
) {
  return useQuery(
    ['cloud', credential?.provider, 'info', credential?.id, { force }],
    () => kaasInfoFetcher(validator, credential, force),
    {
      meta: {
        error: {
          title:
            credential &&
            `Failed to get ${providerTitles[credential.provider]} info`,
          message: '',
        },
      },
      enabled: !!credential && credential.provider === provider,
      retry: 1,
    }
  );
}

async function kaasInfoFetcher<T extends KaasInfo>(
  validator: (info: KaasInfo) => info is T,
  credential?: Credential | null,
  force = false
) {
  if (!credential) {
    return null;
  }

  const info = await getKaasInfo(credential, force);

  return validator(info) ? info : null;
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
