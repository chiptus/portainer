import axios, { parseAxiosError } from '@/portainer/services/axios';
import { Credential } from '@/react/portainer/settings/sharedCredentials/types';
import { Environment } from '@/react/portainer/environments/types';

import { ProvisionOption } from '../WizardK8sInstall/types';

import { KaasInfoResponse, CreateClusterPayload } from './types';
import { parseKaasInfoResponse } from './converter';

function buildUrl(provider: ProvisionOption, action: string) {
  return `/cloud/${provider}/${action}`;
}

export async function createKaasEnvironment(
  payload: CreateClusterPayload,
  provider: ProvisionOption
) {
  try {
    const { data } = await axios.post<Environment>(
      buildUrl(provider, 'cluster'),
      payload
    );
    return data;
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}

export async function getKaasInfo(
  provider: ProvisionOption,
  { id }: Credential,
  force?: boolean
) {
  try {
    const { data } = await axios.get<KaasInfoResponse>(
      buildUrl(provider, 'info'),
      { params: { credentialId: id, force } }
    );
    return parseKaasInfoResponse(data);
  } catch (e) {
    throw parseAxiosError(e as Error);
  }
}
