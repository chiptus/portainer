import { useQueryClient, useMutation } from 'react-query';

import { notifySuccess } from '@/portainer/services/notifications';
import axios, { parseAxiosError } from '@/portainer/services/axios';

import { createKaasEnvironment } from '../WizardKaaS/kaas.service';
import {
  CreateClusterPayload,
  TestSSHConnectionResponse,
} from '../WizardKaaS/types';

import { K8sDistributionType, ProvisionOption } from './types';

export function useInstallK8sCluster() {
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
        notifySuccess('Success', 'K8s installation started');
        client.cancelQueries(['cloud', provider, 'info']);
        return client.invalidateQueries(['environments']);
      },
      meta: {
        error: {
          title: 'Failure',
          message: 'Unable to create K8s environment',
        },
      },
    }
  );
}

export function useTestSSHConnection() {
  return useMutation(
    ({ payload }: { payload: CreateClusterPayload }) =>
      testSSHConnection(payload),
    {
      meta: {
        error: {
          title: 'Failed to test SSH connection',
          message: 'Unable to test SSH connection',
        },
      },
    }
  );
}

async function testSSHConnection(payload: CreateClusterPayload) {
  try {
    const { data } = await axios.post<TestSSHConnectionResponse>(
      `/cloud/${K8sDistributionType.MICROK8S}/cluster`,
      payload,
      { params: { testssh: true } }
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
