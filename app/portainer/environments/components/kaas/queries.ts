import { useMutation, useQuery, useQueryClient } from 'react-query';

import {
  getKaasInfo,
  createKaasEnvironment,
} from '@/portainer/environments/environment.service/kaas';
import { Credential } from '@/portainer/settings/cloud/types';
import { success as notifySuccess } from '@/portainer/services/notifications';

import { KaasCreateFormValues } from './kaas.types';

export function useCloudProviderOptions(credential?: Credential) {
  return useQuery(
    ['cloud', credential?.provider, 'info', credential?.id],
    () => getKaasInfo(credential?.provider, credential?.id),
    {
      meta: {
        error: {
          title: 'Failure',
          message: `Failed to get ${credential?.provider} info`,
        },
      },
      enabled: !!credential?.provider,
      staleTime: 10000,
    }
  );
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
      meta: {
        error: {
          title: 'Failure',
          message: 'Unable to create KaaS environment',
        },
      },
    }
  );
}
