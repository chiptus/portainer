import { useQueryClient, useMutation } from 'react-query';

import { notifySuccess } from '@/portainer/services/notifications';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError } from '@/react-tools/react-query';

import { createKaasEnvironment } from '../WizardKaaS/kaas.service';
import { CreateClusterPayload } from '../WizardKaaS/types';
import { K8sDistributionType } from '../../../types';

import { MicroK8sInfo, ProvisionOption, AddonOption } from './types';

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
      ...withError('Unable to create K8s environment'),
    }
  );
}

export async function getMicroK8sInfo() {
  try {
    const { data } = await axios.get<MicroK8sInfo>(
      `/cloud/${K8sDistributionType.MICROK8S}/info`
    );
    return parseInfoResponse(data);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

function parseInfoResponse(response: MicroK8sInfo): MicroK8sInfo {
  const kubernetesVersions = response.kubernetesVersions.map((v) =>
    buildOption(v.value, v.label)
  );
  const availableAddons = response.availableAddons.map((v) => {
    const a = buildOption(v.value, v.label) as AddonOption;
    a.versionAvailableFrom = v.versionAvailableFrom;
    a.versionAvailableTo = v.versionAvailableTo;
    a.type = v.type;
    return a;
  });

  return {
    kubernetesVersions,
    availableAddons,
    requiredAddons: response.requiredAddons,
  };
}

function buildOption(value: string, label?: string): Option<string> {
  return { value, label: label ?? value };
}

export interface Option<T extends string | number> {
  value: T;
  label: string;
}
