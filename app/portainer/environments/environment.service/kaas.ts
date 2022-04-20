import axios, { parseAxiosError } from '@/portainer/services/axios';

import {
  KaasCreateFormValues,
  KaasInfoResponse,
  KaasProvider,
  KaasProvisionAPIPayload,
} from '../components/kaas/kaas.types';
import { parseKaasInfoResponse } from '../kaas.converter';

function buildUrl(provider: KaasProvider, action: string) {
  return `/cloud/${provider}/${action}`;
}

export async function createKaasEnvironment(values: KaasCreateFormValues) {
  try {
    const payload: KaasProvisionAPIPayload = {
      Name: values.name,
      NodeSize: values.nodeSize,
      NodeCount: values.nodeCount,
      KubernetesVersion: values.kubernetesVersion,
      Region: values.region,
      NetworkID:
        values.type === KaasProvider.CIVO ? values.networkId : undefined,
    };

    await axios.post(buildUrl(values.type, 'cluster'), payload);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

export async function getKaasInfo(provider?: KaasProvider) {
  if (provider) {
    try {
      const { data } = await axios.get<KaasInfoResponse>(
        buildUrl(provider, 'info')
      );
      return parseKaasInfoResponse(data);
    } catch (e) {
      throw parseAxiosError(e as Error);
    }
  }
  return null;
}
