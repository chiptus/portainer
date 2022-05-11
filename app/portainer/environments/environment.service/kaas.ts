import axios, { parseAxiosError } from '@/portainer/services/axios';
import { KaasProvider } from '@/portainer/settings/cloud/types';

import {
  KaasCreateFormValues,
  KaasInfoResponse,
  KaasProvisionAPIPayload,
} from '../components/kaas/kaas.types';
import { parseKaasInfoResponse } from '../kaas.converter';

function buildUrl(
  provider: KaasProvider,
  action: string,
  queryParams?: string
) {
  if (queryParams) {
    return `/cloud/${provider}/${action}?${queryParams}`;
  }
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
      CredentialID: values.credentialId,
    };

    await axios.post(buildUrl(values.type, 'cluster'), payload);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

export async function getKaasInfo(provider?: KaasProvider, id?: number) {
  if (provider) {
    try {
      const { data } = await axios.get<KaasInfoResponse>(
        buildUrl(provider, 'info', `credentialId=${id}`)
      );
      return parseKaasInfoResponse(data);
    } catch (e) {
      throw parseAxiosError(e as Error);
    }
  }
  return null;
}
