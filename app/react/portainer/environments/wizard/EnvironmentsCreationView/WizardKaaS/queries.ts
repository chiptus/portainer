import { useMutation, useQuery, useQueryClient } from 'react-query';

import {
  Credential,
  CredentialType,
  providerToCredentialTypeMap,
} from '@/react/portainer/settings/sharedCredentials/types';
import { success as notifySuccess } from '@/portainer/services/notifications';

import { ProvisionOption } from '../WizardK8sInstall/types';

import { getKaasInfo, createKaasEnvironment } from './kaas.service';
import { CreateClusterPayload, KaasInfo } from './types';

export function useCloudProviderOptions<T extends KaasInfo>(
  provider: ProvisionOption,
  validator: (info: KaasInfo) => info is T,
  credential?: Credential | null,
  force = false
) {
  return useQuery(
    ['cloud', provider, 'info', credential?.id, { force }],
    () => kaasInfoFetcher(provider, validator, credential, force),
    {
      meta: {
        error: {
          title: credential && `Failed to get ${provider} info`,
          message: '',
        },
      },
      enabled:
        !!credential &&
        credential.provider === providerToCredentialTypeMap[provider],
      retry: 1,
    }
  );
}

async function kaasInfoFetcher<T extends KaasInfo>(
  provider: ProvisionOption,
  validator: (info: KaasInfo) => info is T,
  credential?: Credential | null,
  force = false
) {
  if (!credential) {
    return null;
  }

  if (credential.provider === CredentialType.SSH) {
    return null;
  }

  const info = await getKaasInfo(provider, credential, force);

  return validator(info) ? info : null;
}

export function useCreateCluster() {
  const client = useQueryClient();
  return useMutation(
    ({
      payload,
      provider,
    }: {
      payload: CreateClusterPayload;
      provider: ProvisionOption;
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
